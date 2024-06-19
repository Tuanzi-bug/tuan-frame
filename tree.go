package tuan_frame

import (
	"fmt"
	"regexp"
	"strings"
)

type nodeType int

const (
	// 静态路由
	nodeTypeStatic = iota
	// 正则路由
	nodeTypeReg
	// 路径参数路由
	nodeTypeParam
	// 通配符路由
	nodeTypeAny
)

type node struct {
	path string

	// 子节点 path->node
	children map[string]*node
	// 命中路由的执行逻辑
	handler  HandlerFunc
	nodeType nodeType

	// 通配符 * 表达的节点，任意匹配
	starChild *node
	// : 参数路由节点
	paramChild *node
	paramName  string

	// 正则表达式
	regChild *node
	regExpr  *regexp.Regexp
}

// childOf 用于查找子节点
func (n *node) childOf(path string) (*node, bool) {
	if n.children == nil {
		return n.childOfNonstatic(path)
	}
	child, ok := n.children[path]
	if !ok {
		return n.childOfNonstatic(path)
	}
	return child, ok
}

// childOfNonstatic 用于查找非静态路由 // 优先级：正则路由 > 参数路由 > 通配符
func (n *node) childOfNonstatic(path string) (*node, bool) {
	// 正则路由处理
	if n.regChild != nil {
		if n.regChild.regExpr.MatchString(path) {
			return n.regChild, true
		}
	}
	// 参数路由处理
	if n.paramChild != nil {
		return n.paramChild, true
	}
	// 通配符处理
	if n.starChild != nil {
		return n.starChild, true
	}
	return nil, false
}

// insert 前缀树插入节点
func (n *node) insert(path string) *node {
	// 通配符处理
	if path == "*" {
		if n.paramChild != nil {
			panic(fmt.Sprintf("web: 非法路由，已有路径参数路由。不允许同时注册通配符路由和参数路由 [%s]", path))
		}
		if n.regChild != nil {
			panic(fmt.Sprintf("web: 非法路由，已有正则路由。不允许同时注册通配符路由和正则路由 [%s]", path))
		}
		if n.starChild == nil {
			n.starChild = &node{path: path, nodeType: nodeTypeAny}
		}
		return n.starChild
	}
	// 参数路由处理
	if path[0] == ':' {
		param, expr, isReg := n.parseParam(path)
		if isReg {
			// 正则路由添加
			return n.insertRegChild(path, expr, param)
		} else {
			// 参数路由添加
			return n.insertParamChild(path, param)
		}
	}
	if n.children == nil {
		n.children = make(map[string]*node)
	}
	child, ok := n.children[path]
	// 不存在则创建
	if !ok {
		child = &node{
			path:     path,
			nodeType: nodeTypeStatic,
		}
		n.children[path] = child
	}
	return child
}

// insertParamChild 用于插入参数路由子节点
func (n *node) insertParamChild(path string, paramName string) *node {
	if n.regChild != nil {
		panic(fmt.Sprintf("web: 非法路由，已有正则路由。不允许同时注册正则路由和参数路由 [%s]", path))
	}
	if n.starChild != nil {
		panic(fmt.Sprintf("web: 非法路由，已有通配符路由。不允许同时注册通配符路由和参数路由 [%s]", path))
	}
	if n.paramChild != nil {
		//判断是否存在重复可能性
		if n.paramChild.path != path {
			panic(fmt.Sprintf("web: 路由冲突，参数路由冲突，已有 %s，新注册 %s", n.paramChild.path, path))
		}
	} else {
		n.paramChild = &node{path: path, paramName: paramName, nodeType: nodeTypeParam}
	}
	return n.paramChild
}

// insertRegChild 用于插入正则路由子节点
func (n *node) insertRegChild(path string, expr string, paramName string) *node {
	if n.starChild != nil {
		panic(fmt.Sprintf("web: 非法路由，已有通配符路由。不允许同时注册通配符路由和正则路由 [%s]", path))
	}
	if n.paramChild != nil {
		panic(fmt.Sprintf("web: 非法路由，已有路径参数路由。不允许同时注册正则路由和参数路由 [%s]", path))
	}
	if n.regChild != nil {
		if n.regChild.regExpr.String() != expr || n.paramName != paramName {
			panic(fmt.Sprintf("web: 路由冲突，正则路由冲突，已有 %s，新注册 %s", n.regChild.regExpr.String(), expr))
		}
	} else {
		regExpr, err := regexp.Compile(expr)
		if err != nil {
			panic(fmt.Sprintf("web: 正则表达式错误 %s", expr))
		}
		n.regChild = &node{path: path, regExpr: regExpr, paramName: paramName, nodeType: nodeTypeReg}
	}
	return n.regChild
}

// parseParam 用于解析判断是不是正则表达式
// 第一个返回值是参数名字
// 第二个返回值是正则表达式
// 第三个返回值为 true 则说明是正则路由
func (n *node) parseParam(path string) (string, string, bool) {
	// 去除：
	path = path[1:]
	// 规定以()来包含正则表达式 例子：/reg/:id(.*)
	segs := strings.SplitN(path, "(", 2)
	if len(segs) == 2 {
		expr := segs[1]
		if strings.HasSuffix(expr, ")") {
			return segs[0], expr[:len(expr)-1], true
		}
	}
	// 简单参数路由
	return path, "", false
}
