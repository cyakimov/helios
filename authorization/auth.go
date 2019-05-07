package authorization

import (
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
	"time"
)

type Helios struct {
	cel         cel.Env
	expressions []cel.Program
}

func NewAuthorization(expressions []string) *Helios {
	env, err := cel.NewEnv(cel.Declarations(
		decls.NewIdent("request.host", decls.String, nil),
		decls.NewIdent("request.path", decls.String, nil),
		decls.NewIdent("request.ip", decls.String, nil),
		decls.NewIdent("request.time", decls.Timestamp, nil),
	))
	if err != nil {
		log.Fatal(err)
	}

	programs := make([]cel.Program, 0, len(expressions))
	for _, exp := range expressions {
		ast, _ := env.Parse(exp)
		p, err := env.Program(ast)
		if err != nil {
			log.Fatal("Invalid CEL expression %q", err)
		}

		programs = append(programs, p)
	}

	return &Helios{
		cel:         env,
		expressions: programs,
	}
}

func (h *Helios) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		context := getContext(r)

		for _, exp := range h.expressions {
			out, _, err := exp.Eval(context)
			if err != nil {
				log.Fatal(err)
			}

			if out.Value() == false {
				w.WriteHeader(http.StatusForbidden)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func getContext(r *http.Request) map[string]interface{} {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Error(err)
	}
	return map[string]interface{}{
		"request.host": r.Host,
		"request.path": r.RequestURI,
		"request.ip":   ip,
		"request.time": time.Now().Unix(),
	}
}
