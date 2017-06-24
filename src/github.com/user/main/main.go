/*
 *
 * Created by : Akshat Prakash, Freshman, CMU SCS
 * Date: 2/26/17
 * Simple Aadhaar database project for Sphere
 *

Pending Additions:
- Create Query Report Page with locations if possible
- Same Name UID cannot be entered
- Add more info fields for mysql database, just show current 3
- Document code
- Create Readme.md
**/

package main

import (
       "net/http"
       "html/template"
       "path"
       "strings"
       "fmt"
       "encoding/json"
       "database/sql"
        _ "github.com/go-sql-driver/mysql"
       )

type Info struct{
  Id uint8
  Name string
  Aadhaar string
  Phone string
  Email string
  Date string
}

type Profile struct{
  Profile_list []Info
  Aadhaar_error string
}

type Json_response struct{
  UID string
  Name string
  Phone string
  Email string
  Date string
}

type query_data struct{
  Id uint8
  Query_name string
}

func main() {
 
  http.HandleFunc("/database/", handle)
  /*
  http://sphere_aadhaar_database-akshatp694432.codeanyapp.com/database 
  is the weblink
  */
  
  /*API that retreives user info based on Aadhaar*/
  //Query format: https://sphere_aadhaar_database-akshatp694432.codeanyapp.com/query/AADHAAR_NO
  http.HandleFunc("/query/", api_handle)
  
  /*Table that lists the queries made till date*/
  /*
  http://sphere_aadhaar_database-akshatp694432.codeanyapp.com/query_data/
  is the weblink
  */
  http.HandleFunc("/query_data/", query_info_handle)
  
  http.ListenAndServe(":3000", nil)
  //uses default mux 'nil'
}

func query_info_handle(w http.ResponseWriter, req *http.Request){
  db, err := sql.Open("mysql", "root:sphere@/aadhaar_db")
  check_error(err)
  
  query_table := []query_data{}
  rows, err := db.Query("SELECT * FROM queries")
    //query all rows of db
  check_error(err)

    for rows.Next() {//for each row in db
         
         var id uint8
         var query string
         err = rows.Scan(&id, &query)
         //scan datafields in each row
         check_error(err) 
         query_table = append(query_table, query_data{id, query})
         //append next sql record to list of profiles
   }
  
  fp := path.Join("templates","queries.html")//create filepath
  templ, err := template.ParseFiles(fp) //parse and store html template
  check_error(err)
  
  err1 := templ.Execute(w, query_table)//execute html template
  check_error(err1)
  
  err2 := db.Close()
  check_error(err2)
}

func api_handle(w http.ResponseWriter, req *http.Request){
  query := strings.TrimLeft(req.URL.Path,"/query/")
  //Query format: https://sphere_aadhaar_database-akshatp694432.codeanyapp.com/query/AADHAAR_NO
  //open database connection
  db, err := sql.Open("mysql", "root:sphere@/aadhaar_db")
  check_error(err)
  
  var name string
  var phone string
  var email string
  var date string
  //to retrieve data connected AADHAAR number
  if err1 := db.QueryRow("SELECT name,phone,email,date FROM aadhaar WHERE uid = ?", query).Scan(&name,&phone,&email,&date); err1 == sql.ErrNoRows{
    name = "Not in database" 
  } else if err1 == nil{
    fmt.Println(name)
  } else {
  check_error(err1)
  }
  
 
  stmt, err :=  db.Prepare("INSERT INTO queries (query_name) VALUES (?)")
  check_error(err) //safe method to counter mysql injection
  stmt.Exec(name) //insert new record
  
  //create JSON encoding and write it to webpage
  json_response := Json_response{query, name, phone, email, date}
  json.NewEncoder(w).Encode(json_response)
  
  err1 := db.Close()
  check_error(err1)
}

//http handler function for /database
func handle(w http.ResponseWriter, req *http.Request){
  
  req.ParseForm() //parse the aadhaar sign_up from
  profile_base := Profile{} //create an empty holder for aadhaar profiles
              
  //open database connection
  db, err := sql.Open("mysql", "root:sphere@/aadhaar_db")
  check_error(err)

  if req.Method == "POST"{
    insert_record(db, req, &profile_base)
    //if user chooses valid UID, insert record into database
  } 

  //display database data to page
  retreieve_database_records(db, &profile_base) 
  
  //close database connection
  err1 := db.Close();
  check_error(err1)
  
  //display html template
  display_html_template(w, &profile_base)
  
}  

/**********************************HELPER FUNCTIONS********************************/

/* standard error handler 
* REQUIRES: an error type error
* ENSURES: raises error
*/
func check_error(err error){
  if err != nil{
    panic(err)
  }
}

/* checks if given aadhaar id exists in mysql db 
* REQUIRES: database type *sql.DB, 
            uid type string to be checked
* ENSURES: true if uid already exists in database,
           false otws
*/
func aadhaar_exists(db *sql.DB, new_aadhaar string) bool{  
  
  var q_uid string
  if err1 := db.QueryRow("SELECT uid FROM aadhaar WHERE uid = ?",new_aadhaar).Scan(&q_uid); err1 == sql.ErrNoRows{
    return false
  } else {
    check_error(err1)
  }
  return true
}

/* inserts record into mysql database
* REQUIRES: database type *sql.DB,
            http request type *http.Request
            profile type struct pointer *Profile
* ENSURES: inserts record if user chooses a unique
           Aadhaar ID
*/
func insert_record(db *sql.DB, req *http.Request, pfb *Profile){
    new_name := req.Form["name"][0] //retreive name
    new_aadhaar := req.Form["f3"][0]+req.Form["s3"][0]+req.Form["l4"][0]
    new_phone := req.Form["phone"][0]
    new_dob := req.Form["dob"][0]
    new_email := req.Form["email"][0] 
    //retreive uid number
    
    if !aadhaar_exists(db, new_aadhaar){
      stmt, err :=  db.Prepare("INSERT INTO aadhaar (name, uid, phone, email, date) VALUES (?,?,?,?,?)")
      check_error(err) //safe method to counter mysql injection
      stmt.Exec(new_name, new_aadhaar, new_phone, new_email, new_dob) //insert new record
    } else {
        pfb.Aadhaar_error = "*Aadhaar Number Exists. Choose Again"
        //if uid chosen is not unique, raise error
    }  
}

/* retreives records in database
* REQUIRES: a database type *sql.DB,
            profile type struct pointer *Profile
* ENSURES: retreives the records in the mysql db and
           stores in struct Profile
*/
func retreieve_database_records(db *sql.DB, pfb *Profile){
    rows, err := db.Query("SELECT * FROM aadhaar")
    //query all rows of db
    check_error(err)

    for rows.Next() {//for each row in db
         var pid uint8
         var name string
         var uid string
         var phone string
         var email string
         var dob string
         err = rows.Scan(&pid, &name, &uid, &phone, &email, &dob)
         //scan datafields in each row
         check_error(err) 
         pfb.Profile_list = append(pfb.Profile_list, 
                                   Info{pid, 
                                        name, 
                                        uid,
                                        phone,
                                        email,
                                        dob})
         //append next sql record to list of profiles
   }

}

/* displayes html content
* REQUIRES: an http response writer,
            w of type http.ResponseWriter
            and a pointer to struct pfb of
            type Profile
* ENSURES: displays HTML content on webpage
*/
func display_html_template(w http.ResponseWriter, pfb *Profile){
  fp := path.Join("templates","index.html")//create filepath
  templ, err := template.ParseFiles(fp) //parse and store html template
  check_error(err)
  
  err1 := templ.Execute(w, pfb)//execute html template
  check_error(err1)
}
/************************************************************************/

//func api_handle(w http.ResponseWriter, req *http.Request){}


