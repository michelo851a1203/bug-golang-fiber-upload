## unit testing on fiber error
```
I build a fiber app to upload image file and convert to base64 string to response 
and build a unit test to test if status code is okay and response as I thought
but after "go test -v -run TestUploadHandler"  then testing is stuck.
Maybe that is my problem or that is kind of bug. hope guys do me a favor.
And give me better suggestion 
```
- step 1 : 
```sh
go run main.go
```
```sh
curl -X POST -F "file=@{my file path}" http://localhost:8080  
```
> this will be okay will show some base64 string and that is  okay
- step 2 :
```sh
go test -v -run TestUploadHandler
```
> testing somehow get stuck

I'm not sure if that is a kind of bug or I miss something





