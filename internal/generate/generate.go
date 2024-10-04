package generate

import (
	"context"
	"io"
	"net/url"
	"os"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

var (
	// chromeCtx context.Context
	// cancel    context.CancelFunc
	// Fixed-size pool of chromedp contexts
	contextPool chan context.Context
	cancelPool  chan context.CancelFunc
	poolSize    = 2 // Set the desired pool size here

)

func InitChrome() {
	contextPool = make(chan context.Context, poolSize)
	cancelPool = make(chan context.CancelFunc, poolSize)

	for i := 0; i < poolSize; i++ {
		ctx, cnc := chromedp.NewContext(context.Background())
		contextPool <- ctx
		cancelPool <- cnc
	}
}

func ShutdownChrome() {
	close(contextPool)
	close(cancelPool)
	for cancel := range cancelPool {
		cancel()
	}
}

func Generate(htmlData []byte, output io.Writer) error {
	// Acquire a context and cancel function from the pool
	ctx := <-contextPool
	defer func() {
		// Release the context and cancel function back to the pool
		contextPool <- ctx
	}()

	// Create a temporary file to store the HTML content
	tmpFile, err := os.CreateTemp("", "*.html")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name()) // Clean up the temporary file afterward

	// Decode the URL-encoded HTML data in a memory-efficient way
	decodedHTMLData, err := url.QueryUnescape(string(htmlData))
	if err != nil {
		return err
	}

	// Write the HTML content to the temporary file
	if _, err := tmpFile.Write([]byte(decodedHTMLData)); err != nil {
		return err
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}

	// capture pdf
	if err := chromedp.Run(ctx, printToPDF("file://"+tmpFile.Name(), output)); err != nil {
		return err
	}

	return nil
}

func printToPDF(urlstr string, res io.Writer) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.ActionFunc(func(ctx context.Context) error {
			buf, _, err := page.PrintToPDF().WithPaperHeight(11).WithPaperWidth(8.5).WithPrintBackground(false).Do(ctx)
			if err != nil {
				return err
			}

			_, err = res.Write(buf)
			if err != nil {
				return err
			}
			buf = []byte{}
			return nil
		}),
	}
}
