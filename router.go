package tuan_frame

import (
	"fmt"
	"strings"
)

type router struct {
	// 按照HTTP方法进行划分
	trees map[string]*node
}

func newRouter() *router {
	return &router{trees: make(map[string]*node)}
}

// addRoute 注册路由。
// method 是 HTTP 方法
// - 已经注册了的路由，无法被覆盖。例如 /user/home 注册两次，会冲突
// - path 必须以 / 开始并且结尾不能有 /，中间也不允许有连续的 /
// - 不能在同一个位置注册不同的参数路由，例如 /user/:id 和 /user/:name 冲突
// - 不能在同一个位置同时注册通配符路由和参数路由，例如 /user/:id 和 /user/* 冲突
// - 同名路径参数，在路由匹配的时候，值会被覆盖。例如 /user/:id/abc/:id，那么 /user/123/abc/456 最终 id = 456
func (r *router) addRoute(method string, path string, handler HandlerFunc) {
	// 前置条件检查
	if path == "" {
		panic("web: 路由是空字符串")
	}

	if path[0] != '/' {
		panic("web: 路由必须以 / 开头")
	}

	if path != "/" && path[len(path)-1] == '/' {
		panic("web: 路由不能以 / 结尾")
	}
	// 根据请求方法拿到根节点
	root, ok := r.trees[method]
	// 如果不存在，则创建根节点
	if !ok {
		root = &node{path: "/"}
		r.trees[method] = root
	}
	// 解决根节点的问题。原因：strings.Split(path[1:], "/") 得出 len == 1
	if path == "/" {
		if root.handler != nil {
			panic(fmt.Sprintf("web: 路由冲突[%s]", path))
		}
		root.handler = handler
		return
	}
	segs := strings.Split(path[1:], "/")
	for _, s := range segs {
		if s == "" {
			panic(fmt.Sprintf("web: 非法路由。不允许使用 //a/b, /a//b 之类的路由, [%s]", path))
		}
		root = root.insert(s)
	}
	if root.handler != nil {
		panic(fmt.Sprintf("web: 路由冲突[%s]", path))
	}
	root.handler = handler
}

func (r *router) getRoute(method string, path string) (*node, map[string]string, bool) {
	root, ok := r.trees[method]
	if !ok {
		return nil, nil, false
	}
	if path == "/" {
		return root, nil, true
	}
	params := make(map[string]string)
	segs := strings.Split(strings.Trim(path, "/"), "/")
	for _, s := range segs {
		var child *node
		child, ok := root.childOf(s)
		if !ok {
			if root.nodeType == nodeTypeAny {
				return root, params, true
			}
			return nil, nil, false
		}
		if child.paramName != "" {
			params[child.paramName] = s
		}
		root = child
	}
	return root, params, true
}
