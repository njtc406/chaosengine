./protoc --go_out=./datadefine ./datadefineproto/*.proto
./protoc-go-inject-tag -input=./datadefine/*.pb.go

