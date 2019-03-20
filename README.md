# cmake-tools
This package provides automatic configuration of following tools for cmake project:
* clang-format
* clang-tidy

## Prerequisite
```bash
sudo snap install go
```



# Usage
```bash
go get -u -v github.com/Maverobot/cmake-tools

cd $GOPATH/src/github.com/Maverobot/cmake-tools
go build -o cmake-tools main.go

./cmake-tools -path path/to/your/project/CMakeLists.txt
```
