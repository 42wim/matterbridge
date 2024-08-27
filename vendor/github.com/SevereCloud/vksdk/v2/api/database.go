package api // import "github.com/SevereCloud/vksdk/v2/api"

import (
	"github.com/SevereCloud/vksdk/v2/object"
)

// DatabaseGetChairsResponse struct.
type DatabaseGetChairsResponse struct {
	Count int                 `json:"count"`
	Items []object.BaseObject `json:"items"`
}

// DatabaseGetChairs returns list of chairs on a specified faculty.
//
// https://dev.vk.com/method/database.getChairs
func (vk *VK) DatabaseGetChairs(params Params) (response DatabaseGetChairsResponse, err error) {
	err = vk.RequestUnmarshal("database.getChairs", &response, params)
	return
}

// DatabaseGetCitiesResponse struct.
type DatabaseGetCitiesResponse struct {
	Count int                   `json:"count"`
	Items []object.DatabaseCity `json:"items"`
}

// DatabaseGetCities returns a list of cities.
//
// https://dev.vk.com/method/database.getCities
func (vk *VK) DatabaseGetCities(params Params) (response DatabaseGetCitiesResponse, err error) {
	err = vk.RequestUnmarshal("database.getCities", &response, params)
	return
}

// DatabaseGetCitiesByIDResponse struct.
type DatabaseGetCitiesByIDResponse []object.DatabaseCity

// DatabaseGetCitiesByID returns information about cities by their IDs.
//
// https://dev.vk.com/method/database.getCitiesByID
func (vk *VK) DatabaseGetCitiesByID(params Params) (response DatabaseGetCitiesByIDResponse, err error) {
	err = vk.RequestUnmarshal("database.getCitiesById", &response, params)
	return
}

// DatabaseGetCountriesResponse struct.
type DatabaseGetCountriesResponse struct {
	Count int                 `json:"count"`
	Items []object.BaseObject `json:"items"`
}

// DatabaseGetCountries returns a list of countries.
//
// https://dev.vk.com/method/database.getCountries
func (vk *VK) DatabaseGetCountries(params Params) (response DatabaseGetCountriesResponse, err error) {
	err = vk.RequestUnmarshal("database.getCountries", &response, params)
	return
}

// DatabaseGetCountriesByIDResponse struct.
type DatabaseGetCountriesByIDResponse []object.BaseObject

// DatabaseGetCountriesByID returns information about countries by their IDs.
//
// https://dev.vk.com/method/database.getCountriesByID
func (vk *VK) DatabaseGetCountriesByID(params Params) (response DatabaseGetCountriesByIDResponse, err error) {
	err = vk.RequestUnmarshal("database.getCountriesById", &response, params)
	return
}

// DatabaseGetFacultiesResponse struct.
type DatabaseGetFacultiesResponse struct {
	Count int                      `json:"count"`
	Items []object.DatabaseFaculty `json:"items"`
}

// DatabaseGetFaculties returns a list of faculties (i.e., university departments).
//
// https://dev.vk.com/method/database.getFaculties
func (vk *VK) DatabaseGetFaculties(params Params) (response DatabaseGetFacultiesResponse, err error) {
	err = vk.RequestUnmarshal("database.getFaculties", &response, params)
	return
}

// DatabaseGetMetroStationsResponse struct.
type DatabaseGetMetroStationsResponse struct {
	Count int                           `json:"count"`
	Items []object.DatabaseMetroStation `json:"items"`
}

// DatabaseGetMetroStations returns the list of metro stations.
//
// https://dev.vk.com/method/database.getMetroStations
func (vk *VK) DatabaseGetMetroStations(params Params) (response DatabaseGetMetroStationsResponse, err error) {
	err = vk.RequestUnmarshal("database.getMetroStations", &response, params)
	return
}

// DatabaseGetMetroStationsByIDResponse struct.
type DatabaseGetMetroStationsByIDResponse []object.DatabaseMetroStation

// DatabaseGetMetroStationsByID returns information about one or several metro stations by their identifiers.
//
// https://dev.vk.com/method/database.getMetroStationsById
func (vk *VK) DatabaseGetMetroStationsByID(params Params) (response DatabaseGetMetroStationsByIDResponse, err error) {
	err = vk.RequestUnmarshal("database.getMetroStationsById", &response, params)
	return
}

// DatabaseGetRegionsResponse struct.
type DatabaseGetRegionsResponse struct {
	Count int                     `json:"count"`
	Items []object.DatabaseRegion `json:"items"`
}

// DatabaseGetRegions returns a list of regions.
//
// https://dev.vk.com/method/database.getRegions
func (vk *VK) DatabaseGetRegions(params Params) (response DatabaseGetRegionsResponse, err error) {
	err = vk.RequestUnmarshal("database.getRegions", &response, params)
	return
}

// DatabaseGetSchoolClassesResponse struct.
type DatabaseGetSchoolClassesResponse [][]interface{}

// DatabaseGetSchoolClasses returns a list of school classes specified for the country.
//
// BUG(VK): database.getSchoolClasses bad return.
//
// https://dev.vk.com/method/database.getSchoolClasses
func (vk *VK) DatabaseGetSchoolClasses(params Params) (response DatabaseGetSchoolClassesResponse, err error) {
	err = vk.RequestUnmarshal("database.getSchoolClasses", &response, params)
	return
}

// DatabaseGetSchoolsResponse struct.
type DatabaseGetSchoolsResponse struct {
	Count int                     `json:"count"`
	Items []object.DatabaseSchool `json:"items"`
}

// DatabaseGetSchools returns a list of schools.
//
// https://dev.vk.com/method/database.getSchools
func (vk *VK) DatabaseGetSchools(params Params) (response DatabaseGetSchoolsResponse, err error) {
	err = vk.RequestUnmarshal("database.getSchools", &response, params)
	return
}

// DatabaseGetUniversitiesResponse struct.
type DatabaseGetUniversitiesResponse struct {
	Count int                         `json:"count"`
	Items []object.DatabaseUniversity `json:"items"`
}

// DatabaseGetUniversities returns a list of higher education institutions.
//
// https://dev.vk.com/method/database.getUniversities
func (vk *VK) DatabaseGetUniversities(params Params) (response DatabaseGetUniversitiesResponse, err error) {
	err = vk.RequestUnmarshal("database.getUniversities", &response, params)
	return
}
