package profiles

type Profile struct {
    ID               string `json:"id"`
    Name             string `json:"name"`
    Group            string `json:"group"`
    Host             string `json:"host"`
    Port             int    `json:"port"`
    Username         string `json:"username"`
    AuthType         string `json:"authType"`
    PrivateKeyPath   string `json:"privateKeyPath"`
    UseKeyring       bool   `json:"useKeyring"`
    KnownHostsPolicy string `json:"knownHostsPolicy"`
}
