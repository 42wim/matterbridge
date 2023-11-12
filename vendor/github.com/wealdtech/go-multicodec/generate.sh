#!/bin/bash 

OUTPUT=codecs.go

cat >${OUTPUT} <<EOSTART
// Copyright © 2019 Weald Technology Trading
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package multicodec

// codec is the structure holding codec information
type codec struct {
    id uint64
    tag string
    name string
}

var codecs map[string]*codec
var reverseCodecs map[uint64]*codec

func init() {
    codecs = make(map[string]*codec)
    reverseCodecs = make(map[uint64]*codec)

EOSTART
while read line
do
  CODECNAME=$(echo $line |awk -F, '{print $1}' | sed -e 's/^ *//' -e 's/ *$//')
  CODECTAG=$(echo $line |awk -F, '{print $2}' | sed -e 's/^ *//' -e 's/ *$//')
  CODECID=$(echo $line |awk -F, '{print $3}' | sed -e 's/^ *//' -e 's/ *$//')
  CODECDESC=$(echo $line |awk -F, '{print $4}' | sed -e 's/^ *//' -e 's/ *$//')
cat >>${OUTPUT} <<EOCODEC
    codecs["${CODECNAME}"] = &codec{id: ${CODECID},name:"${CODECNAME}",tag:"${CODECTAG}"}
EOCODEC
  # Do not add reverse lookup for deprecated items
  if [[ ! ${CODECDESC} =~ .*deprecated.* ]]
  then
    cat >>${OUTPUT} <<EOCODEC
    reverseCodecs[${CODECID}] = codecs["${CODECNAME}"]
EOCODEC
  fi
done < <(wget -q -O - https://raw.githubusercontent.com/multiformats/multicodec/master/table.csv | tail +2)

cat >>${OUTPUT} <<EOEND
}
EOEND

gofmt -w ${OUTPUT}
