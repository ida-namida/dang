package httpserver

import (
    "dang/service"
    "dang/view"
    "encoding/json"
    "fmt"
    "net/http"

    "github.com/julienschmidt/httprouter"
)

type Server struct {
    httpPort        string
    decisionService service.DecisionService
}

func NewServer(httpPort string, decisionService service.DecisionService) Server {
    return Server{httpPort, decisionService}
}

func (s Server) Start() {
    router := httprouter.New()
    router.POST("/decisions", s.saveDecisions)

    http.ListenAndServe(":" + s.httpPort, router)
}

func (s Server) saveDecisions(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    decoder := json.NewDecoder(r.Body)

    var request view.SaveDecisionRequest
    err := decoder.Decode(&request)
    if err != nil {
        w.Header().Add("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadRequest)
        w.Write([]byte(fmt.Sprintf(`{
            "error": "Cannot parse request into JSON - %s"
        }`, err.Error())))
        return
    }

    err = s.decisionService.SaveDecisions(request)
    if err != nil {
        w.Header().Add("Content-Type", "application/json")
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte(fmt.Sprintf(`{
            "error": "%s"
        }`, err.Error())))
        return
    }

    w.WriteHeader(http.StatusCreated)
}