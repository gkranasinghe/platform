package main

import (
	"net/http"
	"strings"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/logger"
	"github.com/tidepool-org/platform/store"
	"github.com/tidepool-org/platform/user"
	"github.com/tidepool-org/platform/version"

	"github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/ant0ine/go-json-rest/rest"
)

const (
	missingPermissionsError = "missing required permissions"

	dataservicesName = "dataservices"
	useridParamName  = "userid"
)

var (
	validateToken user.ChainedMiddleware
	getPermissons user.ChainedMiddleware
	dataStore     store.Store
	log           = logger.Log.GetNamed(dataservicesName)
)

func initMiddleware() {
	userClient := user.NewServicesClient()
	userClient.Start()
	validateToken = user.NewAuthorizationMiddleware(userClient).ValidateToken
	getPermissons = user.NewPermissonsMiddleware(userClient).GetPermissons
}

func main() {

	log.Info(version.String)

	initMiddleware()

	dataStore = store.NewMongoStore(dataservicesName)

	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	api.Use(&rest.GzipMiddleware{})

	router, err := rest.MakeRouter(
		rest.Get("/version", getVersion),
		rest.Get("/data/:userid/:datumid", validateToken(getPermissons(getData))),
		rest.Post("/dataset/:userid", validateToken(getPermissons(postDataset))),
		rest.Get("/dataset/:userid", validateToken(getPermissons(getDataset))),
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)
	log.Fatal(http.ListenAndServe(":8077", api.MakeHandler()))
}

func checkPermisson(r *rest.Request, expected user.Permission) bool {
	//userid := r.PathParam("userid")
	if permissions := r.Env[user.PERMISSIONS].(*user.UsersPermissions); permissions != nil {

		log.Info("perms found ", permissions)

		//	perms := permissions[userid]
		//	if perms != nil && perms[""] != nil {
		return true
		//	}
	}
	return false
}

func logRequest(r *rest.Request) {
	log.AddTrace("todo")
	log.Info(r.BaseUrl())
	log.Info(r.ContentLength)
	log.Info(r.PathParams)
}

func getVersion(w rest.ResponseWriter, r *rest.Request) {
	w.WriteJson(version.String)
}

func postDataset(w rest.ResponseWriter, r *rest.Request) {

	log.AddTrace(r.PathParam(useridParamName))

	if checkPermisson(r, user.Permission{}) {

		var dataSet data.GenericDataset
		var processedDataset struct {
			Dataset []interface{} `json:"Dataset"`
			Errors  string        `json:"Errors"`
		}

		log.Info("processing")

		err := r.DecodeJsonPayload(&dataSet)

		if err != nil {
			rest.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data, err := data.NewTypeBuilder().BuildFromDataSet(dataSet)
		processedDataset.Dataset = data

		if err != nil {
			rest.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		//TODO: should this be a bulk insert?
		for i := range data {
			err := dataStore.Save(data[i])
			if err != nil {
				rest.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		w.WriteJson(&processedDataset)
		return
	}
	rest.Error(w, missingPermissionsError, http.StatusUnauthorized)
	return
}

func getDataset(w rest.ResponseWriter, r *rest.Request) {

	log.AddTrace(r.PathParam(useridParamName))

	if checkPermisson(r, user.Permission{}) {

		var foundDataset struct {
			data.GenericDataset `json:"Dataset"`
			Errors              string `json:"Errors"`
		}

		userid := r.PathParam(useridParamName)
		log.Info(useridParamName, userid)

		types := strings.Split(r.URL.Query().Get("type"), ",")
		subTypes := strings.Split(r.URL.Query().Get("subType"), ",")
		start := r.URL.Query().Get("startDate")
		end := r.URL.Query().Get("endDate")

		log.Info("params", types, subTypes, start, end)

		var dataSet data.GenericDataset
		err := dataStore.ReadAll(store.IDField{Name: "userId", Value: userid}, &dataSet)

		if err != nil {
			foundDataset.Errors = err.Error()
		}
		foundDataset.GenericDataset = dataSet

		w.WriteJson(&foundDataset)
		return
	}
	rest.Error(w, missingPermissionsError, http.StatusUnauthorized)
	return
}

func getData(w rest.ResponseWriter, r *rest.Request) {

	log.AddTrace(r.PathParam(useridParamName))

	if checkPermisson(r, user.Permission{}) {
		var foundDatum struct {
			data.GenericDatam `json:"Datum"`
			Errors            string `json:"Errors"`
		}

		userid := r.PathParam(useridParamName)
		datumid := r.PathParam("datumid")

		log.Info("userid and datum", userid, datumid)

		foundDatum.GenericDatam = data.GenericDatam{}
		foundDatum.Errors = ""

		w.WriteJson(&foundDatum)
		return
	}
	rest.Error(w, missingPermissionsError, http.StatusUnauthorized)
	return
}
