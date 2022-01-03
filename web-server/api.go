package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/flybot/data-steward-go/common/token"
	"github.com/go-chi/chi/v5"
)

type GetDataRequest struct {
	Rid        int         `json:"rid"`
	Parameters interface{} `json:"params"`
}

type SetDataRequest struct {
	Table string      `json:"table"`
	Where interface{} `json:"where"`
	Data  interface{} `json:"data"`
}

type CreateTaskRequest struct {
	Type    string `json:"task_type"`
	Payload string `json:"payload"`
}
type EmailTaskPayload struct {
	// ID for the email recipient.
	UserID int
}

func ApiRouter() chi.Router {
	r := chi.NewRouter()

	r.Get("/data", GetData)
	r.Post("/data", SetData)
	//r.Post("/new_task", AddTask)
	return r
}

func GetData(w http.ResponseWriter, r *http.Request) {
	var input GetDataRequest

	payload := r.Context().Value("tokenPayload").(*token.Payload)

	rid, _ := strconv.Atoi(r.URL.Query().Get("rid"))
	input.Rid = rid

	params := r.URL.Query().Get("params")
	json.Unmarshal([]byte(params), &input.Parameters)

	paramsString, err1 := json.Marshal(input.Parameters)
	if err1 != nil {
		log.Printf("API GetData marshal params error: %v", err1)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var result string
	err2 := srv.db.GetConnection().Get(&result, fmt.Sprintf("select get_data from extractor.get_data(%d::int, %d::int, '%s')", payload.ID, input.Rid, paramsString))
	if err2 != nil {
		log.Printf("API GetData error: %v", err2)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", result)
}
func SetData(w http.ResponseWriter, r *http.Request) {
	var input SetDataRequest

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		log.Printf("API SetData parse params error: %v", err)
		// If the structure of the body is wrong, return an HTTP error
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	dataString, err1 := json.Marshal(input.Data)
	if err1 != nil {
		log.Printf("API SetData marshal data error: %v", err1)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	whereString, err2 := json.Marshal(input.Where)
	if err1 != nil {
		log.Printf("API SetData marshal where clause error: %v", err2)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	result, err3 := srv.db.GetConnection().Exec(fmt.Sprintf("select set_data from extractor.set_data('%s', '%s', '%s')", input.Table, whereString, dataString))
	if err3 != nil {
		log.Printf("API SetData error: %v", err3)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	rowsAffected, err4 := result.RowsAffected()
	if err4 != nil {
		log.Printf("API SetData error: %v", err4)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "{\"rows\":%d}", rowsAffected)
}

/*func AddTask(w http.ResponseWriter, r *http.Request) {
	cid := r.Context().Value("userID")
	userID := int(cid.(float64))

	var input CreateTaskRequest

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		log.Printf("API CreateTask parse params error: %v", err)
		// If the structure of the body is wrong, return an HTTP error
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     app.Config.Redis.Host,
		Password: app.Config.Redis.Password,
		DB:       app.Config.Redis.Db,
	})

	payload, err := json.Marshal(EmailTaskPayload{UserID: userID})
	if err != nil {
		log.Fatal(err)
	}
	t1 := asynq.NewTask("email:welcome", payload)

	t2 := asynq.NewTask("email:reminder", payload)

	info, err := client.Enqueue(t1)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf(" [*] Successfully enqueued task: %+v", info)

	info, err = client.Enqueue(t2, asynq.ProcessIn(2*time.Minute))
	if err != nil {
		log.Fatal(err)
	}
	log.Printf(" [*] Successfully enqueued task: %+v", info)

	app.Notify("tasker", fmt.Sprintf("New task %s from user %d", input.Type, userID))

}*/
