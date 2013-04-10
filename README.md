go_api_sample
=============

A sample api written in go.

The api does the following
- listens on a port continuously
- accepts parameters through query string 
- processes database , url params 
- returns a json of desired format 

Before starting this API create a file named config.toml similar to config.toml.example
with valid configuration setting values.


E.g.

http://localhost:8080/properties/search.js?current_location_zip=94102&featured_property_ids[]=83&
featured_property_ids[]=101&page_number=1&per_page=24&filters[max_price1]=5475&filters[min_price]=0&
filters[max_beds1]=3&filters[min_beds]=3&filters[max_baths1]=2&filters[min_baths]=0&filters[walkscore]=0&
filters[transitscore]=0&filters[shoppingscore]=0&filters[finedining]=0&filters[artandculture]=0&
filters[schoolrating]=0&filters[kidsfriendly]=0&filters[petsfriendly]=0&filters[cats]=0&filters[dogs]=0&
scroll_flag=0

will take params from query string and query the database and return  a json result
