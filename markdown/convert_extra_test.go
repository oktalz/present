package markdown

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

//revive:disable:line-length-limit

func TestConvertCenter(t *testing.T) {
	source := ".center\nhello\n.center.end"
	expected := `<div style="text-align:center; display: flex; justify-content: center; align-items: center;">hello</div>`
	actual, err := Convert(source)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestConvertImage(t *testing.T) {
	source := ".image(https://example.com/image.png 100px:200px)"
	expected := `<img src="https://example.com/image.png" style="object-fit: contain; width: 100px; height: 200px;">`
	actual, err := Convert(source)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestConvertBoxIcon(t *testing.T) {
	source := ".bx{bxl-docker}"
	expected := `<i class='bx bxl-docker'></i>`
	actual, err := Convert(source)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestConvertTabs(t *testing.T) {
	source := ".tabs\n.tab{active} Tab 1\nContent 1\n.tab Tab 2\nContent 2\n.tabs.end"
	// This is a simplified expected output. The actual output will contain ULIDs.
	expectedStart := `<div class="tab"><button class="tablinks active"`
	expectedMiddle := `>Tab 1</button><button class="tablinks"`
	expectedEnd := `>Tab 2</button></div><div class="tabcontent" id=`
	actual, err := Convert(source)
	assert.NoError(t, err)
	assert.Contains(t, actual, expectedStart)
	assert.Contains(t, actual, expectedMiddle)
	assert.Contains(t, actual, expectedEnd)
}

func TestConvertLink(t *testing.T) {
	source := ".link{page2}(Click me)"
	expected := `<span onclick="setPageWithUpdate(#link#page2#link#)" style="cursor: pointer;">Click me</span>`
	actual, err := Convert(source)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestConvertStyledSpan(t *testing.T) {
	source := ".{color:red}(hello)"
	expected := `<span style="color:red">hello</span>`
	actual, err := Convert(source)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestConvertStyledDiv(t *testing.T) {
	source := ".div{color:blue}(world)"
	expected := `<div style="color:blue">world</div>`
	actual, err := Convert(source)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}
