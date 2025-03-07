package main

import "time"

// Example usage
func main() {
	//Example usage:
	type responseStruct struct {
		entity         string
		ResponseWriter func(string) string
	}

	routine := NewBulkRoutine[responseStruct](BulkRoutineParams[responseStruct]{
		Routine: func(entities []responseStruct, _ any) error {
			println("Routine called with entities: ", entities)
			// use DataAccess.CreateBulk(entities) here.
			// then call entity.ResponseWriter("200 OK") for each entity.
			// this will send the response back to the client.
			for _, entity := range entities {
				println("Entity: ", entity.entity)
				println("Response write: ", entity.ResponseWriter("200 OK"))
			}
			println("Routine finished.")
			return nil
		},
	})
	time.Sleep(4 * time.Second)

	routine.Insert(responseStruct{
		entity:         "entity1",
		ResponseWriter: func(response string) string { return response },
	})

	routine.Insert(responseStruct{
		entity:         "entity2",
		ResponseWriter: func(response string) string { return response },
	})

	time.Sleep(6 * time.Second)

	routine.Insert(responseStruct{
		entity:         "entity3",
		ResponseWriter: func(response string) string { return response },
	})

	time.Sleep(6 * time.Second)
}
