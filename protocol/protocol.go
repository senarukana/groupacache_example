package protocol

type GetRequest struct {
	Key string
}

type SetRequest struct {
	Key   string
	Value string
}

type SetResult int

type GetResponse struct {
	Value string
}
