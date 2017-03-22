#!/bin/bash
go get golang.org/x/tools/cmd/cover

rm -rf ./coverage/
mkdir coverage
list=$(go list $(glide novendor))

for pkg in ${list}; do
  file_name=$(echo ${pkg} | awk -F/ '{print $NF}')
  go test -coverprofile ./coverage/${file_name}.out ${pkg}

  if [ -e ./coverage/${file_name}.out ]; then
    cat ./coverage/${file_name}.out >> ./coverage/coverage
  fi
done

sed -n '/mode: set/!p' coverage/coverage > coverage/coverage.out
sed -i .raw '1s/^/mode: set\
/' coverage/coverage.out

go tool cover -html=coverage/coverage.out
