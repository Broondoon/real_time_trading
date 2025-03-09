package handlers

import (
	"Shared/entities/entity"
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
	_bulkRoutineRegisterGetByUsername = subfunctions.NewBulkRoutine(&subfunctions.BulkRoutineParams[*UserBulk]{
		Routine: registerUsers,
	})
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
	//fmt.Println(w, "OK")
}

// Register handles user registration.
func Register(w network.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
	log.Println("Register() called by handler in Auth-service.")

	// Decode the JSON body into a User object.
	input, err := user.Parse(data)
	if err != nil {
		RespondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid JSON: %v", err))
		return
	}
	_bulkRoutineRegisterGetByUsername.Insert(&UserBulk{UserEntity: input, ResponseWriter: w})

}

func registerUsers(data *[]*UserBulk, TransferParams any) error {
	log.Println("registering users")
	userMap := make(map[string]*UserBulk)
	usernames := make([]string, len(*data))
	for i, d := range *data {
		username := d.UserEntity.GetUsername()
		if _, ok := userMap[username]; ok {
			RespondError(d.ResponseWriter, http.StatusBadRequest, "Username already exists.")
			continue
		}
		userMap[username] = d
		usernames[i] = username
	}
	_, errorList, err := _authDB.GetByForeignIDBulk("Username", usernames)
	if err != nil {
		log.Println("error getting users: ", err)
		for _, d := range userMap {
			RespondError(d.ResponseWriter, http.StatusInternalServerError, "Internal error")
		}
		return err
	}

	for _, d := range userMap {
		log.Println("checking user: ", d.UserEntity.GetUsername())
		if errCode, exists := errorList[d.UserEntity.GetUsername()]; exists {
			log.Println("User has Error: ", errCode, " for user: ", d.UserEntity.GetUsername(), ". If this is 404, this is desirable.")
			if errorList[d.UserEntity.GetUsername()] == http.StatusNotFound {
				hashedPassword, err := HashPassword(d.UserEntity.GetPassword())
				if err != nil {
					log.Printf("error hashing: %s", err)
					RespondError(d.ResponseWriter, http.StatusInternalServerError, "Error hashing password.")
					continue
				}
				d.UserEntity.SetPassword(hashedPassword)
				_bulkRoutineRegisterCreateUser.Insert(d)
				continue
			} else {
				fmt.Println("Error checking user: ", errCode)
				RespondError(d.ResponseWriter, http.StatusInternalServerError, "Internal error")
				continue
			}
		}
		log.Println("User already exists: ", d.UserEntity.GetUsername())
		RespondError(d.ResponseWriter, http.StatusBadRequest, "Username already exists.")
	}
	return nil
}

func createUser(data *[]*UserBulk, TransferParams any) error {
	log.Println("creating users")
	userMap := make(map[string]*UserBulk)
	usersToCreate := make([]user.UserInterface, len(*data))
	for i, d := range *data {
		usersToCreate[i] = d.UserEntity
		userMap[d.UserEntity.GetUsername()] = d
	}
	users, errorList, err := _authDB.CreateBulk(&usersToCreate)
	if err != nil {
		log.Println("error creating users: ", err)
		for _, d := range *users {
			RespondError(userMap[d.GetUsername()].ResponseWriter, http.StatusInternalServerError, "Internal error")
		}
		return err
	}
	for username, d := range errorList {
		fmt.Println("Error creating user: ", d)
		RespondError(userMap[username].ResponseWriter, http.StatusInternalServerError, "Failed to add user to database.")
	}
	for _, d := range *users {
		_bulkRoutineRegisterCreateWallet.Insert(&UserBulk{UserEntity: d, ResponseWriter: userMap[d.GetUsername()].ResponseWriter})
	}
	return nil
}

func createWallet(data *[]*UserBulk, TransferParams any) error {
	users := make(map[string]*UserBulk, len(*data))
	wallets := make([]wallet.WalletInterface, len(*data))
	for i, d := range *data {
		users[d.UserEntity.GetId()] = d
		wallets[i] = wallet.New(wallet.NewWalletParams{
			NewEntityParams: entity.NewEntityParams{},
			UserID:          d.UserEntity.GetId(),
			Balance:         0.0,
		})
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
	for userId := range errorList {
		_bulkRoutineRegisterRemoveUser.Insert(users[userId])
		RespondError(users[userId].ResponseWriter, http.StatusInternalServerError, "Internal error")
	}
	for _, d := range *newWallets {
		RespondSuccess(users[d.GetUserID()].ResponseWriter, nil)
	}
	return nil
}

func removeUser(data *[]*UserBulk, TransferParams any) error {
	log.Printf("Error creating wallets. We need to delete any users we created for this.\n")
	userIDs := make([]string, len(*data))
	for i, d := range *data {
		userIDs[i] = d.UserEntity.GetId()
	}
	errorList, err := _authDB.DeleteBulk(userIDs)
	if err != nil {
		return err
	}
	for _, d := range errorList {
		fmt.Println("WARNING WARNING: Error deleting user: ", d)
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
				fmt.Println("Error checking user: ", errCode)
				RespondError(d.ResponseWriter, http.StatusBadRequest, "Invalid Credentials.")
				continue
			}
		}
		log.Println("Checking password for user: ", user.GetUsername(), " with password: ", d.UserEntity.GetPassword(), " and hash: ", user.GetPassword())
		if CheckPasswordHash(d.UserEntity.GetPassword(), user.GetPassword()) {
			token, err := GenerateToken(user.GetId())
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
