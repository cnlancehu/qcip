#!/bin/bash

platforms=("darwin/amd64" "darwin/arm64" "freebsd/386" "freebsd/amd64" "freebsd/arm" "linux/386" "linux/amd64" "linux/arm" "linux/arm64" "linux/loong64" "linux/mips" "linux/mipsle" "linux/mips64" "linux/mips64le" "linux/ppc64" "linux/ppc64le" "linux/riscv64" "linux/s390x" "netbsd/386" "netbsd/amd64" "netbsd/arm" "openbsd/386" "openbsd/amd64" "openbsd/arm" "openbsd/arm64" "plan9/386" "plan9/amd64" "plan9/arm" "solaris/amd64" "windows/386" "windows/amd64" "windows/arm" "windows/arm64")

package="qcip"
output="dist"

export GOPATH=$GITHUB_WORKSPACE/go
go get -v -u github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/lighthouse

for platform in "${platforms[@]}"
do
    goos="${platform%/*}"
    goarch="${platform#*/}"
    echo "Building for $goos/$goarch"
    GOOS=$goos GOARCH=$goarch go build -o $output/qcip

    if [ $goos = "windows" ]; then
        mv $output/qcip $output/qcip.exe
        zip -j $output/$package-$goos-$goarch.zip $output/qcip.exe config.json
        rm $output/qcip.exe
    else
        mv $output/qcip qcip
        tar -czf $output/$package-$goos-$goarch.tar.gz qcip config.json
        rm qcip
    fi
done
#!/bin/bash

platforms=("darwin/amd64" "darwin/arm64" "freebsd/386" "freebsd/amd64" "freebsd/arm" "linux/386" "linux/amd64" "linux/arm" "linux/arm64" "linux/loong64" "linux/mips" "linux/mipsle" "linux/mips64" "linux/mips64le" "linux/ppc64" "linux/ppc64le" "linux/riscv64" "linux/s390x" "netbsd/386" "netbsd/amd64" "netbsd/arm" "openbsd/386" "openbsd/amd64" "openbsd/arm" "openbsd/arm64" "plan9/386" "plan9/amd64" "plan9/arm" "solaris/amd64" "windows/386" "windows/amd64" "windows/arm" "windows/arm64")

package="qcip"
output="dist"

for platform in "${platforms[@]}"
do
    goos="${platform%/*}"
    goarch="${platform#*/}"
    echo "Building for $goos/$goarch"
    GOOS=$goos GOARCH=$goarch go build -o $output/qcip

    if [ $goos = "windows" ]; then
        mv $output/qcip $output/qcip.exe
        zip -j $output/$package-$goos-$goarch.zip $output/qcip.exe config.json
        rm $output/qcip.exe
    else
        mv $output/qcip qcip
        tar -czf $output/$package-$goos-$goarch.tar.gz qcip config.json
        rm qcip
    fi
done
cd $output
md5sum * > md5sums.txt