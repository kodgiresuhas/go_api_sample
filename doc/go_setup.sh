sudo apt-get install gcc libc6-dev
sudo apt-get install mercurial
hg clone -u release https://code.google.com/p/go
cd go/src
./all.bash
sudo cp ~/go/bin/* /usr/local/bin/
go get github.com/ziutek/mymysql/thrsafe
go get github.com/ziutek/mymysql/autorc
go get github.com/ziutek/mymysql/godrv
go get github.com/pelletier/go-toml
go get labix.org/v2/mgo
