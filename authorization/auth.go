package authorization

import (
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/interpreter/functions"
	log "github.com/sirupsen/logrus"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
	"net"
	"net/http"
	"time"
)

// Helios represents an authorization service instance
type Helios struct {
	cel         cel.Env
	expressions []cel.Program
}

func inNetwork(clientIP ref.Val, network ref.Val) ref.Val {
	snet, ok := network.Value().(string)
	if !ok {
		return types.False
	}
	sip, ok := clientIP.Value().(string)
	if !ok {
		return types.False
	}

	_, subnet, _ := net.ParseCIDR(snet)
	ip := net.ParseIP(sip)

	if subnet.Contains(ip) {
		return types.True
	}

	return types.False
}

// NewAuthorization creates a new authorization service with a given set of rules
func NewAuthorization(expressions []string) *Helios {
	env, err := cel.NewEnv(cel.Declarations(
		decls.NewIdent("request.host", decls.String, nil),
		decls.NewIdent("request.path", decls.String, nil),
		decls.NewIdent("request.ip", decls.String, nil),
		decls.NewIdent("request.time", decls.Timestamp, nil),
		decls.NewFunction("network",
			decls.NewInstanceOverload("network_string_string", []*exprpb.Type{decls.String, decls.String}, decls.String)),
	))
	if err != nil {
		log.Fatal(err)
	}

	programs := make([]cel.Program, 0, len(expressions))
	for _, exp := range expressions {
		parsed, _ := env.Parse(exp)

		ast, cerr := env.Check(parsed)
		if cerr != nil {
			log.Fatalf("Invalid CEL expression: %s", cerr.String())
		}

		// declare function overloads
		funcs := cel.Functions(
			&functions.Overload{
				Operator: "network",
				Binary:   inNetwork,
			})

		p, err := env.Program(ast, funcs)
		if err != nil {
			log.Fatalf("Error while creating CEL program: %q", err)
		}

		programs = append(programs, p)
	}

	return &Helios{
		cel:         env,
		expressions: programs,
	}
}

// Middleware evaluates authorization rules against a request
func (h *Helios) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debugf("Authorizing request %q", r.URL)
		context := getContext(r)

		for _, exp := range h.expressions {
			out, _, err := exp.Eval(context)
			if err != nil {
				log.Errorf("Error evaluating expression: %v", err)
				w.WriteHeader(http.StatusForbidden)
				return
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
		"request.time": time.Now().UTC().Format(time.RFC3339),
	}
}
