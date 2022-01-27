package zinc

type ZinkErrors requestWrapper

func (e ZinkErrors) Error() string {
	return string(e.data)
}

var (
	UnImplementedEndPoint = ZinkErrors{}
)
