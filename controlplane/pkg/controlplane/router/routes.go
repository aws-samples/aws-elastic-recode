// Copyright 2020 AWS ElasticRecode Solution Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// @author Su Wei <suwei007@gmail.com>

package router

import (
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/mux"

	"github.com/stevensu1977/elasticrecode/pkg/controlplane/config"
	"github.com/stevensu1977/elasticrecode/pkg/controlplane/handlers"
)

//InitRoute 初始化API 各类方法
func InitRoute(addr string, msgInput chan string) {

	r := mux.NewRouter()

	r.HandleFunc("/favicon.ico", handlers.NotFound)
	r.HandleFunc("/api", handlers.APIVersion).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1", handlers.APIVersion).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1/schema", handlers.Schema).Methods("GET", "OPTIONS")

	//jobs api
	r.HandleFunc("/api/v1/jobs", handlers.CreateJob).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/v1/vmaf", handlers.CreateVMAFJob).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/v1/jobs/{jobID}", handlers.GetLogs).Methods("GET", "OPTIONS")

	//logs api
	r.HandleFunc("/api/v1/{userID}/jobs", handlers.GetLogs).Methods("GET", "OPTIONS")

	//worker Queue API
	r.HandleFunc("/api/v1/worker/queues", handlers.GetWorkerQueues).Methods("GET", "OPTIONS")

	//workers deployment API
	r.HandleFunc("/api/v1/worker/deployments", handlers.GetWorkerDeployments).Methods("GET", "OPTIONS")

	r.PathPrefix("/static/").Handler((http.StripPrefix("/static/", handlers.FileServerWithCustom404(http.Dir("./static")))))

	r.HandleFunc("/", handlers.Home)
	r.HandleFunc("/index", handlers.Home)

	//初始化websocket
	hub := handlers.NewHub()
	go hub.Run()
	go hub.Consumer(msgInput)

	r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handlers.ServeWs(hub, w, r)
	})

	//helper endporint
	r.Use(middlewareFunc)

	//CORS控制
	config := config.InitConfig()
	if config.CORS {
		r.Use(accessControlMiddleware)
	}

	http.Handle("/", r)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}

}

func accessControlMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS,PUT")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type")

		if r.Method == "OPTIONS" {
			return
		}

		next.ServeHTTP(w, r)
	})
}

//middlewareFunc 中间件处理逻辑
func middlewareFunc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		log.Debug(r.RequestURI)
		//添加Server,Version
		w.Header().Add("Server", "ElasticRecord/ControlePlane")
		w.Header().Add("Version", "v0.1.1")
		///api调用添加application/json
		if strings.Contains(r.RequestURI, "/api") {
			w.Header().Add("Content-Type", "application/json")
		}

		next.ServeHTTP(w, r)
	})
}
