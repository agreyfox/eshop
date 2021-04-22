## login 
```
curl -XPOST http://127.0.0.1:8089/admin/v1/login -d '{"email":"jihua.gao@gmail.com","password":"qweasdzxc"}' -H "Content-Type: applications/json"
```
get Result
```
{"contents":"{\"Category\":{\"Belongto\":\"select,3\",\"game\":\"input,2\",\"name\":\"input,1\",\"online\":\"bool,3\"},\"Game\":{\"logo\":\"file,3\",\"name\":\"input,1\",\"online\":\"bool,4\",\"sname\":\"input,2\"},\"Order\":{\"bad\":\"bool,3\",\"coupon\":\"input,6\",\"email\":\"input,1\",\"product\":\"list,2\",\"social\":\"input,5\",\"status\":\"input,4\",\"worker\":\"input,7\"},\"Product\":{\"desc\":\"textarea,6\",\"game\":\"select,3\",\"logo\":\"file,7\",\"name\":\"input,1\",\"online\":\"bool,5\",\"price\":\"input,4\",\"stock\":\"input,2\"},\"Server\":{\"category\":\"select,3\",\"game\":\"select,2\",\"name\":\"input,1\",\"online\":\"bool,6\",\"product\":\"select,5\",\"tags\":\"input,4\"}}","data":"eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJhdWQiOm51bGwsImV4cCI6IjIwMjAtMDQtMjNUMTU6NDE6MzkuMjExNjg4ODU3KzA4OjAwIiwiaWF0IjpudWxsLCJpc3MiOm51bGwsImp0aSI6bnVsbCwibmJmIjpudWxsLCJzdWIiOm51bGwsInVzZXIiOiJqaWh1YS5nYW9AZ21haWwuY29tIn0.kWTayD-_z9ndRfgxR0mJTA_4qBAXNgFDdZ4y1_JirZE","msg":"Done","retCode":0}

其中contents json 字符串，表示系统目前使用到的内容定义，界面内容由此生。
data为 toke，需要保存在每个请求中。
```
## get config 

```
curl -XGET http://127.0.0.1:8089/admin/v1/config -H "Content-Type: applications/json" --cookie "lqcms_token=eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJhdWQiOm51bGwsImV4cCI6IjIwMjAtMDQtMTNUMTA6Mzg6MDguNDQ0NjQ4ODUzKzA4OjAwIiwiaWF0IjpudWxsLCJpc3MiOm51bGwsImp0aSI6bnVsbCwibmJmIjpudWxsLCJzdWIiOm51bGwsInVzZXIiOiJqaWh1YS5nYW9AZ21haWwuY29tIn0.3Gm4EKl7TRov7IL7QdENNxDqdVWwaY9W6gjL1eE6Nr0"

Return value : 
{"retCode":0,"message":"ok","data":{"admin_email":"jihua.gao@gmail.com","bind_addr":"localhost","cache_disabled":false,"cors_disabled":false,"http_port":"8089","https_port":"443","log_file":"","log_level":"","name":"EShop For Game Product","zip_disabled":false},"meta":{}}
```

## post config  to save all config

```
curl -XPOST http://127.0.0.1:8089/admin/v1/config -H "Content-Type: applications/json" --cookie "lqcms_token=eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJhdWQiOm51bGwsImV4cCI6IjIwMjAtMDQtMTNUMTA6Mzg6MDguNDQ0NjQ4ODUzKzA4OjAwIiwiaWF0IjpudWxsLCJpc3MiOm51bGwsImp0aSI6bnVsbCwibmJmIjpudWxsLCJzdWIiOm51bGwsInVzZXIiOiJqaWh1YS5nYW9AZ21haWwuY29tIn0.3Gm4EKl7TRov7IL7QdENNxDqdVWwaY9W6gjL1eE6Nr0" -d '{"name":"Gao Ji Hua build site"}'
``` 
this change the config==>name 


## get Contents list by type 
```
curl -XGET http://127.0.0.1:8089/admin/v1/contents?type=Site -H "Content-Type: applications/json"
```
order:desc or asc
status: public or pending 
type: content type like game, gold,category etc.

## create content 
```
curl -XPOST http://127.0.0.1:8089/admin/v1/content?type=Site -H "Content-Type: applications/json" -d '{"name":"得到","owner":"cc"}
```
用id = -1 也表示创建，不输入id表示创建

## update content 
```
curl -POST "http://127.0.0.1:8089/admin/v1/content/update?type=Site&id=4" -H "Content-Type: applications/json" -d '{"owner":"ccc"}'
``` 
注意，这里输入的内容必须完整，否则就是删除中间数据

## delete content 
```
curl -XDELETE "http://127.0.0.1:8089/admin/v1/content?type=Site&id=5" -H "Content-Type: applications/json"
``` 

## get content 
```
curl -XGET "http://127.0.0.1:8089/admin/v1/content?type=Site&id=3" -H "Content-Type: applications/json"
```
参数 id， status （public，pending）
## approve content
```
curl -XPOST "http://127.0.0.1:8089/admin/v1/content/approve?type=Site&id=3" -H "Content-Type: applications/json"
```

## reject content 
```
curl -XPOST "http://127.0.0.1:8089/admin/v1/content/reject?type=Site&id=3" -H "Content-Type: applications/json"
```
## search content
```

```
参数,q 是search text ，status（public，pending）

## get all media library 
```
curl -XGET http://127.0.0.1:8089/admin/v1/files -H "Content-Type: applications/json"

```
order : 'desc','asc'
count : 一次取几个
offset: 第几页

## get media content
curl -XGET http://127.0.0.1:8089/admin/v1/file?id=2 -H "Content-Type: applications/json"
```
```
id: 图片id
w : 图片宽度
h: 图片高度


## delete media content 
```
curl -XDELETE http://127.0.0.1:8089/admin/v1/file?id=4 -H "Content-Type: applications/json"
```

## upload the media 
```
 curl -F 'data=@dms-banner.png' http://127.0.0.1:8089/admin/v1/file 
```
将文件传入到媒体库

传多个文件"ah@189.cn0.0.1:8089/admin/v1/file
```
返回多个文件地址
```
{"data":{"file1":"/api/uploads/2020/04/go.mod","file2":"/api/uploads/2020/04/go.sum"},"msg":"ok","retCode":0}
```

## search upload file 
参数 q 字串 , 
```
curl -XGET http://127.0.0.1:8089/admin/v1/files/search?q=readme -H "Content-Type: applications/json"
```
字段里包含readme的


## recover 申请
```
curl -XPOST "http://127.0.0.1:8089/admin/v1/recover" -H "content-Type: applications/json" -d '{
"email":"jihua.gao@gmail.com"}'
```
给用发送recover 邮件

## recover api
```
curl -XPOST "http://127.0.0.1:8089/admin/v1/recover/key" -H "content-Type: applications/json" -d '{
"email":"jihua.gao@gmail.com","key":"xxx","password":"sdsdsdsd"}'
```
三个参数必须，否则不成功
email,key(从邮件获得),password(需要变成的口令)


## backup system 
```
curl -XPOST http://127.0.0.1:8080/admin/v1/backup?source=search
```
备份系统
source 有4种参数：system,analytics,uploads,search ,分别代表系统数据，分析数据，上传媒体库，以及搜索数据


# Customer api 说明

##  user register
```
curl -XPOST http://127.0.0.1:8080/api/v1/register -d '{"email":"e_raeb@yahoo.com","password":
"abc"}'
```
使用post 来注册，除了email，password 
还可以加上, phone 电话号码，social 社交帐号，以及metadata 其他信息,这些信息将被存入系统

## user login 
```
 curl -XPOST http://127.0.0.1:8080/api/v1/login -d '{"email":"e_raeb@yahoo.com","password":
"abc"}'

```
登陆后若成功，返回jwt token，在body和header 都有，需要手工将toke写入以后的api调用header 中，名字为lqcms_token

{"retCode":0,"data":"eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJhdWQiOm51bGwsImV4cCI6IjIwMjAtMDQtMTBUMTE6Mzg6NTkuNzM0MTEwOTkyKzA4OjAwIiwiaWF0IjpudWxsLCJpc3MiOm51bGwsImp0aSI6bnVsbCwibmJmIjpudWxsLCJzdWIiOm51bGwsInVzZXIiOiJlX3JhZWJAeWFob28uY29tIn0.Oqza0Pzh2Vji1xe-N6cadKyTELZZpjXvCCMVQ6KGXY4","meta":{},"Msg":"Done"}

## recover 发出忘记密码请求
``` 
curl -XPOST http://127.0.0.1:8080/api/v1/forgot -d '{"email":"e_raeb@yahoo.com"}'
```
系统将发送recover 邮件到用户邮箱

## recovery 恢复
```
curl -XPOST http://127.0.0.1:8080/api/v1/forgot -d '{"email":"e_raeb@yahoo.com","key":"xxxx","password":"yyyy"}'
```
用户发出恢复指令，key为恢复口令（来自邮件），password为新口令


## contents 获取content列表
```
curl -XGET http://127.0.0.1:8080/api/v1/contents?type=Site 
```
参数有 type, offset,count,order 


## 获取一个content 
```
curl -XGET "http://127.0.0.1:8080/api/v1/content?type=Site&id=1"
```
参数id,type 必须,
还可以使用slug=site-a705f27f-aa86-4b81-b5e2-8f0e4e7fad59

##  使用 slug 获取一uploads  中的文件
```
curl -XGET "http://127.0.0.1:8080/api/v1/uploads?slug=dms-banner.png-1" --cookie "lqcms_token=eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJhdWQiOm51bGwsImV4cCI6IjIwMjAtMDQtMTBUMjM6MjQ6MzkuNzY1OTg3ODMzKzA4OjAwIiwiaWF0IjpudWxsLCJpc3MiOm51bGwsImp0aSI6bnVsbCwibmJmIjpudWxsLCJzdWIiOm51bGwsInVzZXIiOiJlX3JhZWJAeWFob28uY29tIn0.P37cCSVZV5BwOQg5Q5DTifmULoqDiUhXLdrPo0PpDXo"
```

return
{"data":[{"uuid":"48019f16-6a8c-4261-b650-85146a9c02e3","id":12,"slug":"dms-banner.png-1","timestamp":1586524806000,"updated":1586524806000,"name":"dms-banner.png","path":"/api/uploads/2020/04/dms-banner.png","content_length":133177,"content_type":"application/octet-stream"}]}

##  get picture content , 用于图片显示
```
curl -XGET "http://127.0.0.1:8080/api/v1/pics?id=12" 
```
参数 w,h 表示返回图片的宽和高




type Game struct {
	item.Item

	Name   string `json:"name"`
	Sname  string `json:"sname"`
	Logo   string `json:"logo"`
	Online bool   `json:"online"`
}

type Server struct {
	item.Item

	Name     string   `json:"name"`
	Game     string   `json:"game"`
	Online   bool     `json:"online"`
	Category []string `json:"category"`
	Tags     []string `json:"tags"`
	Product  []string `json:"product"`
}

type Category struct {
	item.Item

	Name     string `json:"name"`
	Game     string `json:"game"`
	Online   bool   `json:"online"`
	Belongto string `json:"belongto"`
}

type Order struct {
	item.Item

	Email     string                   `json:"email"`
	Product   []map[string]interface{} `json:"product"`
	Bad       bool                     `json:"bad"`
	Status    uint                     `json:"status"`
	Total     float32                  `json:"total"`
	Due       uint64                   `json:"due"`
	Ispay     bool                     `json:"ispay"`
	Paytime   uint64                   `json:"paytime"`
	Payerinfo string                   `json:"payerinfo"`
	Worker    string                   `json:"worker"`
	Delivery  uint64                   `json:"delivery"`
	Social    string                   `json:"social"`
	Coupon    string                   `json:"coupon"`
}
