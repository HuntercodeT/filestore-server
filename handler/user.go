package handler

import (
	"errors"
	dblayer "filestore-server/db"
	"filestore-server/util"
	// "fmt"
	"io/ioutil"
	"net/http"
	"time"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

// MyClaims 自定义声明结构体并内嵌jwt.StandardClaims,
// jwt包自带的jwt.StandardClaims只包含了官方字段,
// 我们这里需要额外记录一个username字段，所以要自定义结构体,
// 如果想要保存更多信息，都可以添加到这个结构体中
type MyClaims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

var MySecret = []byte("看不见你的笑")

const TokenExpireDuration = time.Hour * 2

//SignupHandler:处理用户注册请求
func SignupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		data, err := ioutil.ReadFile("./static/view/signup.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(data)
		return
	}
	r.ParseForm()
	username := r.Form.Get("username")
	password := r.Form.Get("password")

	if len(username) < 3 || len(password) < 5 {
		w.Write([]byte("invaild parameter"))
		return
	}
	encPassword := util.Sha1([]byte(password + string(MySecret)))
	suc := dblayer.UserSignup(username, encPassword)

	if suc {
		w.Write([]byte("SUCCESS"))
	} else {
		w.Write([]byte("FAILED"))
	}

}

func SigninHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	if r.Method == http.MethodGet {
		// data, err := ioutil.ReadFile("./static/view/signin.html")
		// if err != nil {
		// 	w.WriteHeader(http.StatusInternalServerError)
		// 	return
		// }
		// w.Write(data)
		http.Redirect(w, r, "/static/view/signin.html", http.StatusFound)
		return
	}
	username := r.Form.Get("username")
	password := r.Form.Get("password")
	encPassword := util.Sha1([]byte(password + string(MySecret)))
	// 1. 校验用户名及密码
	pwdCheck := dblayer.UserSignin(username, encPassword)
	if !pwdCheck {
		w.Write([]byte("FAILED"))
		return
	}
	// 2. 生成访问凭证 token
	ExpiresAt, token, err := GenToken(username)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	// 3. 登录成功后重定向到主页
	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: struct {
			Location string
			Username string
			Token    string
			ExpiresAt int64
		}{
			Location: "http://" + r.Host + "/static/view/home.html",
			Username: username,
			Token:    token,
			ExpiresAt:ExpiresAt,
		},
	}
	w.Write(resp.JSONBytes())
}



// UserInfoHandler ： 查询用户信息
func UserInfoHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 解析请求参数
	r.ParseForm()
	username := r.Form.Get("username")
	
	authHeader := r.Header.Get("Authorization")

	// 2. 验证token是否有效
	if authHeader == "" {
		resp := util.RespMsg{
			Code: 2003,
			Msg:  "请求头中auth为空",
		}
		w.Write(resp.JSONBytes())
		return
	}
	// 按空格分割
	parts := strings.SplitN(authHeader, " ", 2)
	if !(len(parts) == 2 && parts[0] == "Bearer") {
		resp := util.RespMsg{
			Code: 2004,
			Msg:  "请求头中auth格式有误",
		}
		w.Write(resp.JSONBytes())
		return
	}
	// parts[1]是获取到的tokenString，我们使用之前定义好的解析JWT的函数来解析它
	_, err := ParseToken(parts[1])
	// _, err := ParseToken(tokenString)
	// fmt.Println(mc)
	if err != nil {
		resp := util.RespMsg{
			Code: 20045,
			Msg:  "无效的Token",
		}
		w.Write(resp.JSONBytes())
		return
	}

	// 3. 查询用户信息
	user, err := dblayer.GetUserInfo(username)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// 4. 组装并且响应用户数据
	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: user,
	}
	w.Write(resp.JSONBytes())
}

// IsValidToken: token校验
// func IsValidToken(token string) bool {
// 	//  1. 从jwt从取出token
// 	//  2. 依次取出 Header Payload
// }

// KeyExpired checks if a key has expired, if the value of user.SessionState.Expires is 0, it will be ignored.
func KeyExpired(expires int64) bool {
	if expires >= 1 {
		return time.Now().After(time.Unix(expires, 0))
	}

	return false
}

// GenToken 生成JWT
func GenToken(username string) (int64, string, error ) {
	// 创建一个我们自己的声明 Payload
	c := MyClaims{
		username, // 自定义字段
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(TokenExpireDuration).Unix(), // 过期时间
			Issuer:    "filestore",                                // 签发人
		},
	}
	// 使用指定的签名方法创建签名对象 Header
	pretoken := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	// 使用指定的secret签名并获得完整的编码后的字符串token Signature
	token, err := pretoken.SignedString(MySecret)
	return c.ExpiresAt, token ,err
}

// ParseToken 解析JWT
func ParseToken(tokenString string) (*MyClaims, error) {
	// 解析token
	token, err := jwt.ParseWithClaims(tokenString, &MyClaims{}, func(token *jwt.Token) (i interface{}, err error) {
		return MySecret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*MyClaims); ok && token.Valid { // 校验token
		return claims, nil
	}
	return nil, errors.New("invalid token")
}
