@echo off

Rem 生成engine的proto文件
protoc.exe --go_out=../actor  -I=../actor/ ../actor/*.proto
protoc-go-inject-tag.exe -input=../actor/*.pb.go


@Rem 生成game的proto文件
@REM protoc.exe  --go_out=../../game/internal/commdefine/proto/datadefine -I=../../game/internal/commdefine/proto ../../game/internal/commdefine/proto/datadefineproto/*.proto
@REM protoc-go-inject-tag.exe -input=../../game/internal/commdefine/proto/datadefine/*.pb.go

Rem 生成example的proto文件
@REM protoc.exe  --go_out=../../example/proto -I=../../example/proto ../../example/proto/*.proto
@REM protoc-go-inject-tag.exe -input=../../example/proto/*.pb.go

PAUSE