package puzzle

import (
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
)

const (
	SuccessfulEventFetch = `
<!DOCTYPE html>
<html lang="en-us">
<body>
<main>
<article>
<div class="eventlist-event"><a href="/">[2026]</a>          </div>
<div class="eventlist-event"><a href="/2025">[2025]</a>          </div>
<div class="eventlist-event"><a href="/2024">[2024]</a>          </div>
<div class="eventlist-event"><a href="/2023">[2023]</a>          </div>
<div class="eventlist-event"><a href="/2022">[2022]</a>          </div>
<div class="eventlist-event"><a href="/2021">[2021]</a>          </div>
<div class="eventlist-event"><a href="/2020">[2020]</a>          </div>
<div class="eventlist-event"><a href="/2019">[2019]</a>          </div>
<div class="eventlist-event"><a href="/2018">[2018]</a>          </div>
<div class="eventlist-event"><a href="/2017">[2017]</a>          </div>
<div class="eventlist-event"><a href="/2016">[2016]</a>          </div>
<div class="eventlist-event"><a href="/2015">[2015]</a>          </div>
</article>
</main>
</body>
</html>
	`

	EmptyBody = `<html></html>`

	SuccessfulDayFetch = `<!DOCTYPE html>
<html lang="en-us">
<body>
<main>
<pre class="calendar calendar-beckon"><a aria-label="Day 1" href="/2025/day/1" class="calendar-day1">         ____           *                               
________/O___\__________|________________________  <span class="calendar-day"> 1</span> <span class="calendar-mark-complete">*</span><span class="calendar-mark-verycomplete">*</span></a>
<a aria-label="Day 2" href="/2025/day/2" class="calendar-day2">   _______||_________                                   
   | _@__ || _o_    |_ _________________________   <span class="calendar-day"> 2</span> <span class="calendar-mark-complete">*</span><span class="calendar-mark-verycomplete">*</span></a>
<a aria-label="Day 3" href="/2025/day/3" class="calendar-day3">   |_&amp;_%__||_oo__^=_[ \|     _    .. .. ..     |        
                     \_]__--|_|___[]_[]_[]__//_|   <span class="calendar-day"> 3</span> <span class="calendar-mark-complete">*</span><span class="calendar-mark-verycomplete">*</span></a>
<a aria-label="Day 4" href="/2025/day/4" class="calendar-day4">                               ____________//___        
__________________________     | \        // @@|   <span class="calendar-day"> 4</span> <span class="calendar-mark-complete">*</span><span class="calendar-mark-verycomplete">*</span></a>
<a aria-label="Day 5" href="/2025/day/5" class="calendar-day5">|_  ___ | .--.           |     |_[#]_@@__//_@@@|        
|_\_|^|_]_|==|_T_T_T_T_T_|                         <span class="calendar-day"> 5</span> <span class="calendar-mark-complete">*</span><span class="calendar-mark-verycomplete">*</span></a>
<a aria-label="Day 6" href="/2025/day/6" class="calendar-day6"> ||   ____________    _______________________           
_||__/            \_  |      |~             |      <span class="calendar-day"> 6</span> <span class="calendar-mark-complete">*</span><span class="calendar-mark-verycomplete">*</span></a>
<a aria-label="Day 7" href="/2025/day/7" class="calendar-day7">|   |   1  2  3    |  |     / \             |____       
|___]__[]_[]_[]__&lt;&gt;|  |    |H_/|\   \\\\\\  | | |  <span class="calendar-day"> 7</span> <span class="calendar-mark-complete">*</span><span class="calendar-mark-verycomplete">*</span></a>
<a aria-label="Day 8" href="/2025/day/8" class="calendar-day8">                      |&lt;&gt;__|H__|_\__|_____|_[_O_|       
 __________________________                   |    <span class="calendar-day"> 8</span> <span class="calendar-mark-complete">*</span><span class="calendar-mark-verycomplete">*</span></a>
<a aria-label="Day 9" href="/2025/day/9" class="calendar-day9">/      ______         __  |  _________________O__       
[     |(__) [         ] \ |__|  [  ]          | |  <span class="calendar-day"> 9</span> <span class="calendar-mark-complete">*</span><span class="calendar-mark-verycomplete">*</span></a>
<a aria-label="Day 10" href="/2025/day/10" class="calendar-day10">o=====|_____o=========o_|_[__]_____-/_-/_-/___|_|       
_________||______ ______________________________   <span class="calendar-day">10</span> <span class="calendar-mark-complete">*</span><span class="calendar-mark-verycomplete">*</span></a>
<a aria-label="Day 11" href="/2025/day/11" class="calendar-day11">|  ___          | |                            |        
|_|   |_(:::::)_| |   ^      ^      ^      ^   |   <span class="calendar-day">11</span> <span class="calendar-mark-complete">*</span><span class="calendar-mark-verycomplete">*</span></a>
<a aria-label="Day 12" href="/2025/day/12" class="calendar-day12">    |      .--.   |  &lt;^&gt;    &lt;^&gt;    &lt;^&gt;    &lt;^&gt;  |        
    '------'  '---#_&lt;&lt;^&gt;&gt;__&lt;&lt;^&gt;&gt;__&lt;&lt;^&gt;&gt;__&lt;&lt;^&gt;&gt;_|   <span class="calendar-day">12</span> <span class="calendar-mark-complete">*</span><span class="calendar-mark-verycomplete">*</span></a>
</pre>
</main>
</body>
</html>`
)

func htmlResponder(str string) httpmock.Responder {
	return func(r *http.Request) (*http.Response, error) {
		response := httpmock.NewStringResponse(http.StatusOK, str)
		// set header type to text/html
		response.Header.Set("Content-Type", "text/html")

		return response, nil
	}
}

func setupHappyPath() {
	httpmock.RegisterResponder(http.MethodGet, EventsURL, htmlResponder(SuccessfulEventFetch))

	httpmock.RegisterResponder(
		http.MethodGet,
		`=~^https://adventofcode\.com/?$`,
		htmlResponder(EmptyBody),
	)

	httpmock.RegisterResponder(http.MethodGet, `=~^https://adventofcode.com/(\d+)/?$`, func(r *http.Request) (*http.Response, error) {
		response := httpmock.NewStringResponse(http.StatusOK, SuccessfulDayFetch)
		// set header type to text/html
		response.Header.Set("Content-Type", "text/html")
		return response, nil
	})
}

func TestFetchSuccess(t *testing.T) {
	httpmock.Activate(t)
	defer httpmock.Deactivate()

	setupHappyPath()

	fetcher := NewEventFetcher(http.DefaultClient)
	events, err := fetcher.Fetch(true)

	if err != nil {
		t.Error(err)
		return
	}

	if len(events) != 12 {
		t.Errorf("unexpected event length: expected 3, got: %d", len(events))
		return
	}

	if events[0].Year != "2026" {
		t.Errorf("unexpected year, expected: 2026, got: %s", events[0].Year)
		return
	}

	days := events[1].Days

	if len(days) != 12 {
		t.Errorf("unexpected days, expected: 12, got: %d", len(days))
		return
	}

}

func TestYearExtractor(t *testing.T) {
	extracted := YearExtractor("[2026]")

	if extracted != "2026" {
		t.Errorf("unexpected extracted value, expected: 2026, got: %s", extracted)
		return
	}
}
