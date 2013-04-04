package main

import (
	"fmt"
	"net/http"
  "strings"
  "encoding/json"
	"labix.org/v2/mgo" 
	"labix.org/v2/mgo/bson" 
	"strconv"
  "github.com/ziutek/mymysql/mysql"
  _ "github.com/ziutek/mymysql/native" // Native engine
  _ "github.com/ziutek/mymysql/thrsafe" // Thread safe engine
	"math/rand"
	"github.com/pelletier/go-toml"
) 

type MarkerHash struct{
	Latitude float64
	Longitude float64
	Key int
	Title string
	Icon Icon
	SingleInfoWindow bool
}
type Icon struct{
	Image string
	Iconsize []int
}
type Property struct {
	 //Id string
	 Prop_id int 
	 Pet_policy string
	 Walkscore int
	 Neighborhood_id string
	 Amenities  []string
	 Transit_score int
	 School_rating int
	 Shopping_score int 
	 Fine_dining_score int
	 Art_culture_score int
	 Kids_friendly_score int
	 Pets_friendly_score int
	 Min_beds int
	 Max_beds int
	 Min_baths int 
	 Max_baths int
	 Min_sqft int
	 Max_sqft int
	 Min_price int
	 Max_price int
	 Zip string
	 City string
	 State string
	 Lat float64
	 Lng float64
	 Address string
	 Images []string
	 Title string
	 Property_type string
	 //Created_at Time
	 //Updated_at Time
	 Avg_rate float64
	 Processed_amenities []string
	 Price string
	 Baths string
	 Beds string
	 Sqft string
	 Price_per_sqft string
	 Images_count int
	 Image string 
	 Marker_hash_for_google MarkerHash
} 

const lenPath = len("/properties/search.js")

var max_price int
var max_beds int
var max_baths int


func are_filters_applied(query_params map[string][]string) bool{
   for key := range query_params {
      if strings.HasPrefix(key,"filters") {
         return true
      }
   }
   return false
}

func include(search int,featured_property_ids []int) bool{
	for _,v:= range featured_property_ids{
		if v == search{
			return true
		}
	}
	return false
}

func zip_around_10_miles(cur_zip string, limit int) []string {
		if cur_zip == "" {return nil}
		config,_:= toml.LoadFile("config.toml")
	 	mysql_user_name ,_ := config.Get("mysql_user_name").(string)
	 	mysql_password ,_ := config.Get("mysql_password").(string)
		mysql_db_name ,_ := config.Get("mysql_db_name").(string)
	  mysql_url ,_ := config.Get("mysql_url").(string)

    db := mysql.New("tcp", "", mysql_url, mysql_user_name,mysql_password, mysql_db_name)
    err := db.Connect()
    if err != nil {
        panic(err)
    }
    rows,_, err := db.Query("SELECT `zip_distance`.* FROM `zip_distance` WHERE `zip_distance`.`PrimaryZip` = "+cur_zip+" LIMIT 1")
    if err != nil {
        panic(err)
    }
    var zips string
    for _, row := range rows {
        val1 := row[2].([]byte)
        row =row
     		zips = string(val1)
    }
	  return strings.Split(zips,",")
}

func loadResult(r *http.Request )([]Property,[]int,int) {
   params := r.URL.Query()//[lenPath:]
 
	 config,_:= toml.LoadFile("config.toml")
	 mongo_url ,_ := config.Get("mongo_db_url").(string)	
   mongo_db_name ,_ := config.Get("mongo_db_name").(string)
	 mongo_collection_name ,_ := config.Get("mongo_collection_name").(string)

	 session, err := mgo.Dial(mongo_url)
	 if err != nil { 
	  panic(err) 
	 } 
	 defer session.Close() 
	 session.SetMode(mgo.Monotonic, true) 
	 c := session.DB(mongo_db_name).C(mongo_collection_name) 

	 search_results:= []Property{}
	 page := 1
	 if params["page_number"]	  !=nil{
		 page,_= 	strconv.Atoi(params["page_number"][0])
	 }	 
	
	 per_page  := 24
	 if params["per_page"] != nil	{
	 	 per_page,_ = 	strconv.Atoi(params["per_page"][0])
	 }
	 
	 zip_codes  :=	params["filters[zip_codes][]"];
	 
	 var featured_property_ids []int
	 f_prop_ids := params["featured_property_ids[]"]
   for _,v:= range f_prop_ids {
    val ,_:= strconv.Atoi(v)
   	featured_property_ids = append(featured_property_ids,val)
   } 	

	 current_location_zip := ""
	 if params["current_location_zip"]!=nil{
		 current_location_zip = params["current_location_zip"][0]
	 }

 	 scroll_flag :=0
 	 scrl_flag := params["scroll_flag"]
   if scrl_flag != nil {
	   scroll_flag,_ = strconv.Atoi(scrl_flag[0])
   }

	 property_count := 0
	 search_query := bson.M{}
	 var zip_within_10_miles []string
	 if zip_codes == nil || are_filters_applied(params) == false { 
		 //if scroll_flag == 0 {
		 	zip_within_10_miles = zip_around_10_miles(current_location_zip, 50)
		 	//unique
		 //}

		 search_query["zip"] = bson.M{"$in" :  zip_within_10_miles }

     if scroll_flag == 0 && are_filters_applied(params) == false {
			 temp := Property{}
       c.Find(search_query).Sort("-max_price").One(&temp) 
       max_price = temp.Max_price

       c.Find(search_query).Sort("-max_beds").One(&temp) 
       max_beds = temp.Max_beds

       c.Find(search_query).Sort("-max_baths").One(&temp) 
       max_baths = temp.Max_baths

       property_count,_ = c.Find(search_query).Count()
     }
	 }	
	 if are_filters_applied(params) == true{
		 min_price := 0
		 if params["filters[min_price]"] != nil{
		   min_price , _ = strconv.Atoi(params["filters[min_price]"][0])		 
		 }
		 
		 max_price := 0
		 if params["filters[max_price1]"] != nil {
	  	 max_price , _ = strconv.Atoi(params["filters[max_price1]"][0])
  	 }
  	 
  	 min_beds :=0
  	 if params["filters[min_beds]"] != nil {
		   min_beds , _  = strconv.Atoi(params["filters[min_beds]"][0])
		 }

		 max_beds := 0
		 if params["filters[max_price1]"] != nil {	
	  	 max_beds , _  = strconv.Atoi(params["filters[max_price1]"][0])
	   }
	   	
	   min_baths := 0	
	   if params["filters[min_beds]"] != nil { 
		   min_baths , _ = strconv.Atoi(params["filters[min_baths]"][0])
		 }
		 
		 max_baths := 0
		 if params["filters[max_baths1]"] != nil {
	  	 max_baths , _ = strconv.Atoi(params["filters[max_baths1]"][0])
	  }	 
  	 
  	 query := search_query
  	 query["max_price"] = bson.M{"$gte" : min_price }
		 query["min_price"] = bson.M{"$lte" : max_price }
		 query["max_beds"]  = bson.M{"$gte" : min_beds  }
		 query["min_beds"]  = bson.M{"$lte" : max_beds  }
 		 query["max_baths"] = bson.M{"$gte" : min_baths }
 		 query["min_baths"] = bson.M{"$lte" : max_baths } 
 		 
 		 amenities :=params["filters[amenities][]"];
 		 if amenities !=nil {
			 query["amenities"]= bson.M{"$all" :amenities }
 		 }
 		 
		 if zip_codes !=nil {
			 query["zip"]= bson.M{"$in" : zip_codes }
 		 }

  	 neighborhoods_id :=params["filters[neighborhoods_id][]"];
 		 if neighborhoods_id != nil {
			 query["neighborhoods_id"]= bson.M{"$in" : neighborhoods_id }
 		 }

 		 cats := ""
 		 if params["filters[cats]"] != nil {
 		   cats = params["filters[cats]"][0]
 		 }
 		 dogs := ""
 		 if params["filters[dogs]"] != nil {
 		 	 dogs = params["filters[dogs]"][0]	
 		 }
 		 if  cats == "1" && dogs == "1" {
				query["pet_policy"] = "cats and dogs"
 		 }else if cats == "1"{
 				query["pet_policy"] = "cats only"
 		 }else if dogs == "1"{
 				query["pet_policy"] = "dogs only"
 		 }
	
 		 if scroll_flag == 0{
 		 	 property_count ,_ = c.Find(query).Count()
 		 }

  	 walkscore := 0
  	 if params["filters[walkscore]"] != nil {
  	   walkscore,_ = strconv.Atoi(params["filters[walkscore]"][0])
  	 }
  	 
  	 transitscore := 0 
  	 if params["filters[transitscore]"] != nil {
  	   transitscore,_ = strconv.Atoi(params["filters[transitscore]"][0])
  	 }
  	 
  	 schoolrating := 0 
  	 if params["filters[schoolrating]"] != nil {
	   	 schoolrating,_ = strconv.Atoi(params["filters[schoolrating]"][0])
  	 }
  	 
  	 shoppingscore := 0 
  	 if params["filters[shoppingscore]"] != nil {
   	 	 shoppingscore,_ =  strconv.Atoi(params["filters[shoppingscore]"][0])
  	 }
  	 
  	 finedining := 0
  	 if params["filters[finedining]"] != nil {
	  	 finedining,_ = strconv.Atoi(params["filters[finedining]"][0])
	   }
	   
	   artandculture := 0	 
	   if params["filters[artandculture]"] != nil {
	  	 artandculture,_ = strconv.Atoi(params["filters[artandculture]"][0])	   
	   }

  	 kidsfriendly := 0
  	 if params["filters[kidsfriendly]"] != nil {
  	 	 kidsfriendly,_ = strconv.Atoi(params["filters[kidsfriendly]"][0])
  	 }
  	 
  	 petsfriendly := 0 
  	 if params["filters[petsfriendly]"] != nil {
	  	 petsfriendly,_ = strconv.Atoi(params["filters[petsfriendly]"][0])
	   }	 
  	 
  	 total_weight := walkscore + transitscore + schoolrating + shoppingscore + finedining + artandculture + kidsfriendly + petsfriendly
  	 if total_weight != 0{

			 prop_weights := make(map[int]int)
			 result := Property{}
			 iter := c.Find(query).Iter()

			 for iter.Next(&result) {
			 		prop_weights[result.Prop_id] = (walkscore * result.Walkscore + transitscore * result.Transit_score + 
			 		schoolrating * result.School_rating + shoppingscore * result.Shopping_score + finedining * result.Fine_dining_score +
			 		artandculture * result.Art_culture_score + kidsfriendly * result.Kids_friendly_score + 
			 		petsfriendly * result.Pets_friendly_score)/total_weight
       }

       start := ((page - 1)*per_page)
       end1 := start + per_page - 1

       var values,keys []int
 			 for k,v:= range prop_weights {
					keys = append(keys,k)
					values = append(values,v)
			 }

			 for i:=0;i < (len(keys) - 1);i++{
			 	for j:=0;j < (len(keys) -i -1) ;j++{
			 		if values[j] > values[j+1]{
			 			swap := values[j]
			 			values[j] = values[j+1]
			 			values[j+1] = swap
						swap = keys[j]
			 			keys[j] = keys[j+1]
			 			keys[j+1] = swap			 			
			 		}
			 	}
			 }

       keys=keys[start:end1+1]
       if keys !=nil {
				 query["prop_id"]= bson.M{"$in" : keys }
 		 	 }

       c.Find(query).Sort("max_price").All(&search_results)

  	 }else{
			 c.Find(query).Limit(per_page).Sort("max_price").Skip((page-1)*per_page).All(&search_results)
  	 }
  	 return search_results,featured_property_ids,property_count
  }

  search_query_for_count :=search_query
  search_query_for_count["zip"]= current_location_zip
  property_count,_ = c.Find(search_query_for_count).Count()
  if property_count > (page-1)*per_page{

		properties_in_location_zip := search_results
  	c.Find(search_query_for_count).Limit(per_page).Skip((page-1)*per_page).All(&properties_in_location_zip)
  	properties_in_location_zip_count := len(search_results)

		if properties_in_location_zip_count < per_page{
			remaining_prop := per_page - properties_in_location_zip_count

			search_query["zip"] = bson.M{"$ne" : current_location_zip}
			c.Find(search_query_for_count).Limit(per_page).Skip(remaining_prop).Sort("zip").All(&search_results)

			for k := range search_results{
				properties_in_location_zip = append(properties_in_location_zip , search_results[k])
			}
		}
  	return properties_in_location_zip,featured_property_ids,property_count
  }else{
  	offset := (((page-1)*per_page)-property_count)
  	c.Find(search_query_for_count).Limit(per_page).Skip(offset).Sort("zip").All(&search_results)
		return search_results,featured_property_ids,property_count	       	
  }
  return search_results,featured_property_ids,property_count
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
    results,featured_property_ids,property_count := loadResult( r )
		for k := range results {
		  results[k].Processed_amenities = results[k].Amenities
		  
			if results[k].Min_price == results[k].Max_price {
			  results[k].Price = strconv.Itoa( results[k].Min_price)
			}else{
			  results[k].Price = strconv.Itoa( results[k].Min_price) + " - " + strconv.Itoa( results[k].Max_price)
			}
				
			if (results[k].Min_baths == results[k].Max_baths) && results[k].Min_baths > 0{
				results[k].Baths = strconv.Itoa(results[k].Min_baths)
			}else if results[k].Max_baths != 0{
				results[k].Baths = strconv.Itoa( results[k].Min_baths) + " - " + strconv.Itoa( results[k].Max_baths)
			}else{
				results[k].Baths = "1-2"
			}
			
			if results[k].Min_beds == results[k].Max_beds{
				results[k].Beds = strconv.Itoa(results[k].Min_beds)
			}else if results[k].Max_beds >= 0 { 
				results[k].Beds = strconv.Itoa(results[k].Min_beds) + " - " + strconv.Itoa(results[k].Max_beds)
			}else{
				results[k].Beds = "--"
			}
			
			if (results[k].Min_sqft == results[k].Max_sqft) && results[k].Min_sqft >= 0 {
				results[k].Sqft = strconv.Itoa(results[k].Min_sqft)				
			}else if results[k].Max_sqft >= 0 { 
				results[k].Sqft = strconv.Itoa(results[k].Min_sqft) + " - " + strconv.Itoa(results[k].Max_sqft)
			}else{
				results[k].Sqft = "--"
			}
			
			results[k].Price_per_sqft = "--"
			
			results[k].Images_count = len(results[k].Images)
			
			if results[k].Images_count > 0 {
				if results[k].Images_count >= 4 { 
					results[k].Image = results[k].Images[3]
				}else{
					results[k].Image = results[k].Images[rand.Intn(results[k].Images_count)]
				}	
			}
			
			image := ""
			if include(results[k].Prop_id,featured_property_ids)==true 			{
				image = "/assets/featured_property.png"
			}else{
	  		image = "/assets/multi_family_marker.png"
 			}
 			if results[k].Lat != 0{
	 			results[k].Marker_hash_for_google.Icon.Image =image
				results[k].Marker_hash_for_google.Icon.Iconsize = []int{27,40}
	 			results[k].Marker_hash_for_google.Latitude = results[k].Lat
	 			results[k].Marker_hash_for_google.Longitude = results[k].Lng
	 			results[k].Marker_hash_for_google.Key = results[k].Prop_id
	 			results[k].Marker_hash_for_google.Title = results[k].Title
	 			results[k].Marker_hash_for_google.SingleInfoWindow = true
 			}
		}
    result_json, _ := json.Marshal(results)
    w.Header().Set("Content-Type", "application/json")
    fmt.Fprint(w,string(result_json))
    fmt.Println("Total Count",property_count)
	  return
}

func main() {
    http.HandleFunc("/properties/search.js", searchHandler)
    http.ListenAndServe(":8080", nil)
}
