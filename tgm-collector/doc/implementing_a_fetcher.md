Fetchers a how we collect values from the cluster.
After adding commands to the fetcher when `Fetch` is called these commands are executed on the provided context.

If we wanted to fetch the uptime we could simply add the following command


```go
var uptimeFetcher *Fetcher

func init(){
	uptimeFetcher = NewFetcher()

    // "MyUptimeValue" is the key which the value be returned in the result of the fetch
    // "uptime" is the command to execute
    // the final parameter weither to trim whitespace from the output.
    err := uptimeFetcher.AddNewCommand("MyUptimeValue", "uptime", true)
    if err != nil {
        panic("Failed to add uptime command to fetcher")
    }
}
```

We need to provide a stuct to upack the data into. The field which we want to pass the value into must have a `fetcherKey` tag to allow the the fetcher to populate the field. The value of the tag must match the key in the result.

Note: You will need defined a method `GetAnalyserFormat` [see collector documentation](implementing_a_collector.md) for clarification.

```go
type MyUptime struct {
    Raw string `fetcherKey:"MyUptimeValue"`
}

func (uptime *MyUptime) GetAnalyserFormat() {
    formatted := callbacks.AnalyserFormatType{
		ID: "MyUptimeValue",
		Data: []string{
			uptime.Raw,
		},
	}
	return &formatted, nil
}

func GetUptime(ctx clients.ContainerContext) (MyUptime, error) {
    uptime := MyUptime{}
    uptimeFetcher.Fetch(ctx, &uptime)
    if err != nil {
        log.Debugf("failed to fetch uptime %s", err.Error())
		return gpsNav, err
	}
	return uptime, nil
}
```

If you wish to extract/filter/transform the raw data you can define a post processing function as follows
```go
type MyUptime struct {
    Raw      string `fetcherKey:"MyUptimeValue"`
    Uptime   string `fetcherKey:"Uptime"`
    Load1    string `fetcherKey:"load1"`
    Load5    string `fetcherKey:"load5"`
    Load15   string `fetcherKey:"load15"`
}

var (
    uptimeFetcher *Fetcher
    uptimeRegEx = regex.MustCompile(
        `\d+:\d+:\d+ up (.*)` +
        `,\s+\d+:\d+,\s+\d+ users,` +
        `\s+load average:\s+(\d+\.?\d*), (\d+\.?\d*), (\d+\.?\d*)`
    // Example 12:57:08 up 49 days,  1:46,  0 users,  load average: 7.10, 9.33, 10.35
    )
)

func init(){
    uptimeFetcher = NewFetcher()
    err := uptimeFetcher.AddNewCommand("MyUptimeValue", "uptime", true)
    if err != nil {
        panic("Failed to add uptime command to fetcher")
    }
    // Assign processUptime to uptimeFetcher.
    uptimeFetcher.SetPostProcessor(processUptime)
}

// Here we define something to take the raw result from MyUptimeValue and extracts values
func processUptime(result map[string]string) (map[string]string, error) {
    processedResult := make(map[string]string)
    match := uptimeRegEx.FindStringSubmatch(result["MyUptimeValue"])
    if len(match) == 0 {
        return processedResult, fmt.Errorf("unable to to parse uptime output: %s", result["MyUptimeValue"])
    }
    processedResult["Uptime"] = match[1]
    processedResult["load1"] = match[2]
    processedResult["load5"] = match[3]
    processedResult["load15"] = match[4]
    return processedResult, nil
}
```
