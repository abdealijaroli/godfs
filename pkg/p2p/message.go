package p2p

import "encoding/json"

func SerializeMessage(msg Message) ([]byte, error) {
    return json.Marshal(msg)
}

func DeserializeMessage(data []byte) (Message, error) {
    var msg Message
    err := json.Unmarshal(data, &msg)
    return msg, err
}
