package httpclient_test

/*
func ExampleInstrumentedClient() {
	metrics := httpclient.newMetrics("foo", "bar")
	prometheus.DefaultRegisterer.MustRegister(metrics)

	c := httpclient.InstrumentedClient{
		Options:     httpclient.Options{PrometheusMetrics: metrics},
		Application: "test",
	}

	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	if resp, err := c.Do(req); err == nil {
		body, _ := io.ReadAll(resp.Body)
		fmt.Print(string(body))
		_ = resp.Body.Close()
	}
}

func ExampleCacher() {
	metrics := httpclient.newMetrics("foo", "bar")
	prometheus.DefaultRegisterer.MustRegister(metrics)

	table := []httpclient.CacheTableEntry{
		{
			Path: "/foo/.+",
			IsRegExp: true,
			Expiry:   5 * time.Second,
		},
	}

	c := httpclient.NewCacher(nil, "test", httpclient.Options{PrometheusMetrics: metrics}, table, time.Minute, time.Hour)

	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	if resp, err := c.Do(req); err == nil {
		body, _ := io.ReadAll(resp.Body)
		fmt.Print(string(body))
		_ = resp.Body.Close()
	}
}
*/
