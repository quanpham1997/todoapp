package main

import (
	"io"
	"net/http"
	"github.com/gorilla/mux"
	"strconv"
	log "github.com/sirupsen/logrus"
	_ "github.com/go-sql-driver/mysql"
       "github.com/jinzhu/gorm"
    _ "github.com/jinzhu/gorm/dialects/mysql"
	"encoding/json"
)

var db, _ = gorm.Open("mysql","root:root@/todolist?charset=utf8&parseTime=True&loc=Local")

type TodoItemModel struct{
	Id int `gorm:"primary_key"`
	Description string
	Completed bool 
	
}

func CreateItem(w http.ResponseWriter, r *http.Request){ 
	description		:= r.FormValue("description") // obtain the value from the POST operation 
	log.WithFields(log.Fields{"description":description}).Info("Add new TodoItem. Saving to database")
	todo			:= &TodoItemModel{Description: description, Completed : false}
	//We use the value as the description to insert into our database.
	
	db.Create(&todo)	//After that, we create the todo object and persist it in our database
	result			:=db.Last(&todo)
	//// Get last record, order by primary key desc
	// SELECT * FROM users ORDER BY id DESC LIMIT 1;
	

	w.Header().Set("Content-type","application/json") // set your content-type header so clients know to expect json
	json.NewEncoder(w).Encode(result.Value)
	// Lastly, we query the database and return the query 
	//result to the client to make sure the operation is success.
}


func Healthz(w http.ResponseWriter, r *http.Request){ 
	log.Info("API health is OK")
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, `{"alive": true}`) // Healthz function that’ll respond {"alive": true}
								//to the client every time it’s invoked.
}
func init(){
	log.SetFormatter(&log.TextFormatter{})
	log.SetReportCaller(true)
	// set up our logrus logger settings.
}

func UpdateItem(w http.ResponseWriter, r *http.Request){
	// Get URL parameter from mux create a map of route variables 
	vars := mux.Vars(r) 
	id, _ := strconv.Atoi(vars["id"])
	
	// Test if the TodoItem exist in DB 
	err := GetItemById(id)
	if err == false { 
			w.Header().Set("Content-Type", "application/json")
			io.WriteString(w , `{"delete": false, "error": "Record not Found"}`)
		} else{ 
			log.WithFields(log.Fields{"Id":id}).Info("Deleting TodoItem")
			todo := &TodoItemModel{}
			db.First(&todo , id) // Retrieving objects with primary key 
								// // SELECT * FROM todo_item_models WHERE id = _ ;

			db.Delete(&todo)// DELETE from emails where id = 10;

			w.Header().Set("Content-Type", "application/json")
			io.WriteString(w, `{"deleted":true}`)
		}
}








func GetItemById(Id int)bool{ 
	todo := &TodoItemModel{}
	result := db.First(&todo, Id)
	if result.Error != nil{ 
			log.Warn("TodoItem not found in database")
			return false
	}
	return true
}

func GetTodoItems(completed bool) interface{}{ 
	var todos []TodoItemModel
	TodoItems := db.Where("completed = ?", completed).Find(&todos).Value 
	// Slice of primary keys
	//	select * from todo_item_models where completed = 0;
	//+----+--------------+-----------+
	//| id | description  | completed |
	//+----+--------------+-----------+
	//|  1 | Feed the Cat |         0 |
	//|  2 | Feed the Dog |         0 |
	//|  3 | Eat pasta    |         0 |
	//+----+--------------+-----------+
	//

	return TodoItems
}





func GetCompleteItems(w http.ResponseWriter, r *http.Request){ 
	log.Info("Get completed TodoItem")
	completedTodoItems := GetTodoItems(true)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(completedTodoItems)


}


func GetIncompleteItems(w http.ResponseWriter, r *http.Request){ 
	log.Info("Get Incomplete TodoItem")
	IncompleteTodoItems := GetTodoItems(false)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(IncompleteTodoItems)


}








func main() {
	defer db.Close()
	db.Debug().DropTableIfExists(&TodoItemModel{})
	db.Debug().AutoMigrate(&TodoItemModel{})
	log.Info("Starting TodoList API server")



	//This is equivalent to how http.HandleFunc() works: if an incoming request URL matches one of the paths,  
	//the corresponding handler is called passing (http.ResponseWriter, *http.Request) as parameters.	
	router := mux.NewRouter()
	router.HandleFunc("/Healthz", Healthz).Methods("GET")
	router.HandleFunc("/todo",CreateItem).Methods("POST")//, we register the new route /todo with an HTTP POST request into our new CreateItem() function.
	http.ListenAndServe(":8000",router)// We route /healthz HTTP GET requests to the Health() function. The router will listen to port 8000


}























