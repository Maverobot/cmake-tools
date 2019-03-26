# cmake-tools
This package provides automatic configuration of following tools for cmake project:
* clang-format
* clang-tidy

**No dependency** is needed for running this tool with AppImage.

### Usage
```bash
wget https://github.com/Maverobot/cmake-tools/releases/download/continuous/cmake-tools-v0.0.1.glibc2.3.3-x86_64.AppImage -O cmake-tools.AppImage
chmod +x cmake-tools.AppImage
./cmake-tools/cmake-tools.AppImage -path path/to/your/project/CMakeLists.txt

### Compilation
#### Prerequisite
```bash
sudo snap install go
```

#### Download, compile and use

```bash
go get -u -v github.com/Maverobot/cmake-tools

cd $GOPATH/src/github.com/Maverobot/cmake-tools
go build -o cmake-tools main.go

./cmake-tools -path path/to/your/project/CMakeLists.txt
```
