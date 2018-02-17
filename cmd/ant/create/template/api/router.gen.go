
package api

import (
    tp "github.com/henrylee2cn/teleport"
)

// Route registers handlers to router.
func Route(root string, router *tp.Router) {
    // root router group
    rootGroup := router.SubRoute(root)

    // custom router
    customRoute(rootGroup.ToRouter())

    // automatically generated router
    rootGroup.RoutePull(new(ApiHandlers))
}
