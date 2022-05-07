package handler

import (
	"filestore-server/util"
	"log"
	"net/http"
	"strings"
)

// HTTPIntercepto:请求拦截器
// func HTTPInterceptor(h http.HandlerFunc) http.HandlerFunc {
// 	return http.HandleFunc(
// 		func(w http.ResponseWriter, r *http.Request) {
// 			r.ParseForm()
// username := r.Form.Get("username")
// 			h(w,r)
// 		})
// }
func HTTPInterceptor(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			r.ParseForm()
			username := r.Form.Get("username")
			token := r.Form.Get("token")
			if len(token) != 0 {
				claims, err := ParseToken(token)
				// log.Println(claims)
				if err != nil || username != claims.Username {
					resp := util.RespMsg{
						Code: 20045,
						Msg:  "无效的Token",
					}
					http.Redirect(w, r, "/static/view/signin.html", http.StatusForbidden)
					log.Println(resp.JSONString())
					return
				}
			} else {
				authHeader := r.Header.Get("Authorization")
				// fmt.Println(authHeader)
				//验证登录token是否有效
				if authHeader == "" {
					resp := util.RespMsg{
						Code: 2003,
						Msg:  "请求头中auth为空",
					}
					// fmt.Println("22222222222222222")
					http.Redirect(w, r, "/static/view/signin.html", http.StatusForbidden)
					log.Println(resp.JSONString())
					return
				}
				// 按空格分割
				parts := strings.SplitN(authHeader, " ", 2)
				if !(len(parts) == 2 && parts[0] == "Bearer") {
					resp := util.RespMsg{
						Code: 2004,
						Msg:  "请求头中auth格式有误",
					}
					// fmt.Println("333333333333333333")
					http.Redirect(w, r, "/static/view/signin.html", http.StatusForbidden)
					log.Println(resp.JSONString())
					return
				}
				// parts[1]是获取到的tokenString，我们使用之前定义好的解析JWT的函数来解析它
				claims, err := ParseToken(parts[1])
				if username != claims.Username {
					resp := util.RespMsg{
						Code: 2005,
						Msg:  "用户名不匹配，请重新登录",
					}
					http.Redirect(w, r, "/static/view/signin.html", http.StatusForbidden)
					log.Println(resp.JSONString())
					return
				}
				if err != nil {
					resp := util.RespMsg{
						Code: 2006,
						Msg:  "无效的Token",
					}
					http.Redirect(w, r, "/static/view/signin.html", http.StatusForbidden)
					log.Println(resp.JSONString())
					// log.Println(claims)
					// log.Println(username)
					return
				}
			}

			// log.Println(claims)

			h(w, r)
		})
}
