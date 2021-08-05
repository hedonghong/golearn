package framework

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strings"
)

type Engine struct {
	*RouterGroup
	router *httprouter.Router
	groups []*RouterGroup
}

type RouterGroup struct {
	prefix      string //路由分组前缀
	middlewares []HandlerFunc
	parent      *RouterGroup
	engine      *Engine
}

func New() *Engine {
	engine := &Engine {
		router: httprouter.New(),
	}
	engine.RouterGroup = &RouterGroup{
		engine: engine,
	}
	engine.groups = []*RouterGroup{engine.RouterGroup}
	return engine
}

func (group *RouterGroup) Group(prefix string) *RouterGroup {
	engine := group.engine
	newGroup := &RouterGroup{
		prefix: group.prefix + prefix,
		parent: group,
		engine: engine,
	}
	engine.groups = append(engine.groups, newGroup)
	return newGroup
}

func (group *RouterGroup) Use(middlewares ...HandlerFunc) {
	group.middlewares = append(group.middlewares, middlewares...)
}

func (group *RouterGroup) addRoute(method, comp string, handlerFunc httprouter.Handle)  {
	pattern := group.prefix + comp
	group.engine.router.Handle(method, pattern, handlerFunc)
}

func (group *RouterGroup) GET(pattern string, handler httprouter.Handle) {
	group.addRoute("GET", pattern, handler)
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request)  {
	var middlewares []HandlerFunc
	for _, group := range e.groups {
		if strings.HasPrefix(r.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}
	c := &Context{
		router: e.router,
		handlers: middlewares,
		Writer: w,
		Request: r,
		index: -1,
		length: len(middlewares),
	}
	c.Next()
	//e.router.ServeHTTP(w, r)
}

func (e *Engine) Run(addr string)  {
	http.ListenAndServe(addr, e)
}