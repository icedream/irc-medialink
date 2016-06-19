package limitedio

import "io"

type limitedReader struct {
	io.Reader
	rest int
}

func NewLimitedReader(r io.Reader, limit int) io.Reader {
	return &limitedReader{r, limit}
}

func (r *limitedReader) Read(data []byte) (n int, err error) {
	if r.rest <= 0 {
		err = io.EOF
		return
	}

	var dataSize int
	if len(data) < r.rest {
		dataSize = len(data)
	} else {
		dataSize = r.rest
	}

	actualData := make([]byte, dataSize)
	n, err = r.Reader.Read(actualData)
	if n > 0 {
		copy(data, actualData)
	}
	r.rest -= (n)

	return
}

type limitedReadCloser struct {
	*limitedReader
	closeMethod func() error
}

func NewLimitedReadCloser(r io.ReadCloser, limit int) io.Reader {
	return &limitedReadCloser{&limitedReader{r, limit}, r.Close}
}

func (rc *limitedReadCloser) Close() error {
	return rc.closeMethod()
}
