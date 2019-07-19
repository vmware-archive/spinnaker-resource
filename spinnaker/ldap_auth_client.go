package spinnaker

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	neturl "net/url"
	"os"
)

type LdapAuthClient struct {
	username string
	password string
}

func NewLdapAuthClient(u string, p string) *LdapAuthClient {
	return &LdapAuthClient{
		username: u,
		password: p,
	}
}

func blockRedirect(req *http.Request, via []*http.Request) error {
	return http.ErrUseLastResponse
}

func (hac *LdapAuthClient) GetClient(apiUrl string) (*http.Client, error)  {
	cookieJar, _ := cookiejar.New(nil)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := http.Client {
		CheckRedirect: blockRedirect,
		Jar: cookieJar,
		Transport: tr,
	}

	url := fmt.Sprintf("%s/login", apiUrl)
	resp, err := client.PostForm(
		url,
		neturl.Values{
			"username": {hac.username},
			"password": {hac.password},
		},
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "login failed: %s\n", err)
		return nil, err
	}

	if resp.StatusCode != 302 {
		return nil, fmt.Errorf("login status code=%d", resp.StatusCode)
	}

	// At this point, we don't really know if login successfully or not. Because
	// login response doesn't contain the result. So we make an API to test if
	// login succeeded.
	url = fmt.Sprintf("%s/applications", apiUrl)
	resp, err = client.Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 302 {
		return nil, fmt.Errorf("Invalid username or password")
	}

	if (resp.StatusCode != 200) {
		// This should never happen.
		return nil, fmt.Errorf("Unexpected status code %d. This could " +
			"be a transient issue or a bug", resp.StatusCode)
	}

	return &client, nil
}
