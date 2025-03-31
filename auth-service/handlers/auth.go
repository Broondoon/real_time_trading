package handlers

import (
	"Shared/entities/user"
	"Shared/entities/wallet"
	"Shared/network"
	subfunctions "Shared/subfunctions/Multithreading"
	"databaseAccessAuth"
	"databaseAccessUserManagement"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// ---------- Utility Functions ----------

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GenerateToken(userID string) (string, error) {
	var jwtsecret = []byte(os.Getenv("JWT_SECRET"))
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Hour * 1).Unix(),
	})
	return token.SignedString(jwtsecret)
}

// ---------- Response Helpers ----------

func RespondSuccess(w network.ResponseWriter, data interface{}) {
	response := map[string]interface{}{
		"success": true,
		"data":    data,
	}
	w.EncodeResponse(http.StatusOK, response)
}

func RespondError(w network.ResponseWriter, statusCode int, errorMsg string) {
	log.Println("RespondError: ", errorMsg)
	log.Println("RespondErrorCode: ", statusCode)
	response := map[string]interface{}{
		"success": false,
		"error":   errorMsg,
	}
	w.EncodeResponse(statusCode, response)
}

// ---------- Dependency Injection ----------

// authDB is the dependency injected from main.go.
// It implements databaseAccessAdduth.AuthDataAccessInterface.
var _authDB databaseAccessAuth.DatabaseAccessInterface
var _bulkRoutineRegisterGetByUsername subfunctions.BulkRoutineInterface[*UserBulk]
var _bulkRoutineRegisterCreateUser subfunctions.BulkRoutineInterface[*UserBulk]
var _bulkRoutineRegisterCreateWallet subfunctions.BulkRoutineInterface[*UserBulk]
var _bulkRoutineRegisterRemoveUser subfunctions.BulkRoutineInterface[*UserBulk]
var _bulkRoutineLoginGetByUsername subfunctions.BulkRoutineInterface[*UserBulk]

type UserBulk struct {
	UserEntity     user.UserInterface
	ResponseWriter network.ResponseWriter
}

var _networkManager network.NetworkInterface
var _walletAccess databaseAccessUserManagement.WalletDataAccessInterface

// InitializeAuthHandlers sets up the dependency for the handlers.
func InitializeUser(db databaseAccessAuth.DatabaseAccessInterface, networkManager network.NetworkInterface, walletAccess databaseAccessUserManagement.WalletDataAccessInterface) {
	_authDB = db
	_walletAccess = walletAccess
	_bulkRoutineRegisterCreateUser = subfunctions.NewBulkRoutine(&subfunctions.BulkRoutineParams[*UserBulk]{
		Routine: createUser,
	})
	_bulkRoutineRegisterCreateWallet = subfunctions.NewBulkRoutine(&subfunctions.BulkRoutineParams[*UserBulk]{
		Routine: createWallet,
	})
	_bulkRoutineRegisterRemoveUser = subfunctions.NewBulkRoutine(&subfunctions.BulkRoutineParams[*UserBulk]{
		Routine: removeUser,
	})
	_bulkRoutineLoginGetByUsername = subfunctions.NewBulkRoutine(&subfunctions.BulkRoutineParams[*UserBulk]{
		Routine: loginUsers,
	})
	_networkManager = networkManager

	_networkManager.AddHandleFuncUnprotected(network.HandlerParams{Pattern: "authentication/register", Handler: Register})
	_networkManager.AddHandleFuncUnprotected(network.HandlerParams{Pattern: "authentication/login", Handler: Login})
	http.HandleFunc("/health", healthHandler)
}

// ---------- HTTP Handlers ----------

func healthHandler(w http.ResponseWriter, r *http.Request) {
	// Simple check: you might expand this to test database connectivity, etc.
	w.WriteHeader(http.StatusOK)
	//log.Println(w, "OK")
}

// Register handles user registration.
func Register(writer network.ResponseWriter, data []byte, queryParams url.Values, requestType string) {

	// Decode the JSON body into a User object.
	new_user, err := user.Parse(data)
	if err != nil {
		RespondError(writer, http.StatusBadRequest, fmt.Sprintf("Invalid JSON: %v", err))
		return
	}

	// This was pulled out from the deprecated registerUsers() func
	// It just changes the user's password to a hashed one so that passwords are not stored plaintext
	hashedPassword, err := HashPassword(new_user.GetPassword())
	// hashedPassword, err := HashPassword(new_user.UserEntity.GetPassword())
	if err != nil {
		log.Printf("error hashing: %s", err)
		RespondError(writer, http.StatusInternalServerError, "Error hashing password.")
	}
	new_user.SetPassword(hashedPassword)

	/// This bulk routine was replaced in order to skip duplicate username checking.
	// _bulkRoutineRegisterGetByUsername.Insert(&UserBulk{UserEntity: new_user, ResponseWriter: writer})
	_bulkRoutineRegisterCreateUser.Insert(&UserBulk{UserEntity: new_user, ResponseWriter: writer})
}

func createUser(data *[]*UserBulk, TransferParams any) error {
	userMap := make(map[string]*UserBulk)
	usersToCreate := make([]user.UserInterface, len(*data))
	for i, d := range *data {
		usersToCreate[i] = d.UserEntity
		userMap[d.UserEntity.GetUniquePairing().String()] = d
	}
	users, errorList, err := _authDB.CreateBulk(&usersToCreate)
	if err != nil {
		log.Println("error creating users: ", err)
		for _, d := range *users {
			RespondError(userMap[d.GetUniquePairing().String()].ResponseWriter, http.StatusInternalServerError, "Internal error")
		}
		return err
	}
	for _, d := range *users {
		if err, ok := errorList[d.GetUniquePairing().String()]; ok {
			log.Println("Error creating user: ", d)
			// This is where we can inject some individual case-by-case error handling
			if err == http.StatusBadRequest {
				log.Println("Duplicate error; ", err)
				RespondError(userMap[d.GetUniquePairing().String()].ResponseWriter, http.StatusBadRequest, "Duplicate username error")
			} else {
				RespondError(userMap[d.GetUniquePairing().String()].ResponseWriter, http.StatusInternalServerError, "Internal error")
			}
		} else {
			_bulkRoutineRegisterCreateWallet.Insert(&UserBulk{UserEntity: d, ResponseWriter: userMap[d.GetUniquePairing().String()].ResponseWriter})
		}
	}
	return nil
}

func createWallet(data *[]*UserBulk, TransferParams any) error {
	users := make(map[string]*UserBulk, len(*data))
	wallets := make([]wallet.WalletInterface, len(*data))
	for i, d := range *data {
		users[d.UserEntity.GetUniquePairing().String()] = d
		w := wallet.New(wallet.NewWalletParams{
			UserID:  d.UserEntity.GetId(),
			Balance: 0.0,
		})
		w.SetUnqiuePairing(d.UserEntity.GetUniquePairing())
		wallets[i] = w

	}
	newWallets, errorList, err := _walletAccess.CreateBulk(&wallets)
	if err != nil {
		log.Printf("Error creating wallet: %v\n", err.Error())
		for _, d := range *data {
			RespondError(d.ResponseWriter, http.StatusInternalServerError, "Internal error")
		}
		removeUser(data, nil)
		return err
	}
	for _, d := range *newWallets {
		if _, ok := errorList[d.GetUniquePairing().String()]; ok {
			log.Println("Error creating wallet: ", d)
			_bulkRoutineRegisterRemoveUser.Insert(users[d.GetUniquePairing().String()])
			RespondError(users[d.GetUniquePairing().String()].ResponseWriter, http.StatusInternalServerError, "Internal error")
		} else {
			RespondSuccess(users[d.GetUniquePairing().String()].ResponseWriter, nil)
		}
	}
	return nil
}

func removeUser(data *[]*UserBulk, TransferParams any) error {
	log.Printf("Error creating wallets. We need to delete any users we created for this.\n")
	userIDs := make([]*uuid.UUID, len(*data))
	for i, d := range *data {
		userIDs[i] = d.UserEntity.GetId()
	}
	errorList, err := _authDB.DeleteBulk(userIDs)
	if err != nil {
		return err
	}
	for _, d := range errorList {
		log.Println("WARNING WARNING: Error deleting user: ", d)
	}
	return nil
}

// Login handles user login.
func Login(w network.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
	input, err := user.Parse(data)
	if err != nil {
		RespondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid JSON: %v", err))
		return
	}
	_bulkRoutineLoginGetByUsername.Insert(&UserBulk{UserEntity: input, ResponseWriter: w})
}

func loginUsers(data *[]*UserBulk, TransferParams any) error {
	userMap := make(map[string]*UserBulk)
	usernames := make([]string, len(*data))
	for i, d := range *data {
		username := d.UserEntity.GetUsername()
		userMap[username] = d
		usernames[i] = username
	}
	users, errorList, err := _authDB.GetByForeignIDBulk("Username", usernames)
	if err != nil {
		for _, d := range *users {
			RespondError(userMap[d.GetUsername()].ResponseWriter, http.StatusInternalServerError, "Internal error")
		}
		return err
	}

	for _, user := range *users {
		d := userMap[user.GetUsername()]
		if errCode, exists := errorList[user.GetUsername()]; exists {
			log.Println("User has Error: ", errCode)
			if errorList[d.UserEntity.GetUsername()] == http.StatusNotFound {
				RespondError(d.ResponseWriter, http.StatusBadRequest, "Invalid Credentials.")
				continue
			} else {
				log.Println("Error checking user: ", errCode)
				RespondError(d.ResponseWriter, http.StatusBadRequest, "Invalid Credentials.")
				continue
			}
		}
		if CheckPasswordHash(d.UserEntity.GetPassword(), user.GetPassword()) {
			token, err := GenerateToken(user.GetIdString())
			if err != nil {
				RespondError(d.ResponseWriter, http.StatusInternalServerError, "Token generation failed.")
				continue
			}
			RespondSuccess(d.ResponseWriter, map[string]interface{}{"token": token})
		} else {
			RespondError(d.ResponseWriter, http.StatusBadRequest, "Invalid Credentials.")
		}
	}
	return nil
}
