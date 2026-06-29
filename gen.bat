@echo off
protoc --proto_path=proto ^
       --go_out=proto --go_opt=paths=source_relative ^
       --go-grpc_out=proto --go-grpc_opt=paths=source_relative ^
       proto/user/user.proto
         
echo "Proto code generated successfully!"