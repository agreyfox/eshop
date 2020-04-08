## login 
```
curl -XPOST http://127.0.0.1:8089/admin/v1/login -d '{"email":"jihua.gao@gmail.com","password":"qweasdzxc"}' -H "Content-Type: applications/json"
```
get Result
```
lqcms_token=eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJhdWQiOm51bGwsImV4cCI6IjIwMjAtMDQtMTNUMTA6Mzg6MDguNDQ0NjQ4ODUzKzA4OjAwIiwiaWF0IjpudWxsLCJpc3MiOm51bGwsImp0aSI6bnVsbCwibmJmIjpudWxsLCJzdWIiOm51bGwsInVzZXIiOiJqaWh1YS5nYW9AZ21haWwuY29tIn0.3Gm4EKl7TRov7IL7QdENNxDqdVWwaY9W6gjL1eE6Nr0
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

传多个文件
```
curl -F 'file1=@go.mod' -F 'file2=@go.sum'  http://127.0.0.1:8089/admin/v1/file
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
