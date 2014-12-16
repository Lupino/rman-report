package sources

type Stat struct {
    Name string
    Value string
}

type SourceData struct {
    Name string
    Hostname string
    Stats []Stat
}

type SourceHost struct {
    Port int    `json:"port"`
    Host string `json:"host"`
}

type Source interface {
    // Fetches containers or pod information from all the nodes in the cluster.
    // Returns:
    GetInfo() ([]SourceData, error)
}

func NewSource() (Source, error) {
    return newExternalSource()
}
