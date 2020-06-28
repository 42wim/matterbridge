// Package request HTTP client for golang
//  - Make http requests from Golang
//  - Intercept request and response
//  - Transform request and response data
//
// GET
//
//     client := request.Client{
//         URL:    "https://google.com",
//         Method: "GET",
//         Params: map[string]string{"hello": "world"},
//     }
//     resp, err := client.Do()
//
// POST
//
//     client := request.Client{
//         URL:    "https://google.com",
//         Method: "POST",
//         Params: map[string]string{"hello": "world"},
//         Body:   []byte(`{"hello": "world"}`),
//     }
//     resp, err := client.Do()
//
// Content-Type
//
//     client := request.Client{
//         URL:          "https://google.com",
//         Method:       "POST",
//         ContentType: request.ApplicationXWwwFormURLEncoded, // default is "application/json"
//     }
//     resp, err := client.Do()
//
// Authorization
//
//     client := request.Client{
//         URL:       "https://google.com",
//         Method:    "POST",
//         BasicAuth:      request.BasicAuth{
//             Username:"user_xxx",
//             Password:"pwd_xxx",
//         }, // xxx:xxx
//     }
//
//     resp, err := client.Do()
//
// Cookies
//     client := request.Client{
//         URL:       "https://google.com",
//         Cookies:[]*http.Cookie{
//              {
//               Name:  "cookie_name",
//               Value: "cookie_value",
//              },
//         },
//     }
//
//     resp, err := client.Do()
package request
