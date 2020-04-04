# md 文件实时预览
没有用 ```fsnotifiy``` 简单的使用了轮询实现。

# Install
```
go get https://github.com/eavesmy/gmd_preview
cd gmd_preview
go build main.go
```

# Usage
```
./main --file test.md
```

# Params
```
--file  [required] 
--port  [options]  [default 8080]
--css   [options]  [default https://eva7base.com/css/sspai.css]
--width [options]  [default 80vw]
```
