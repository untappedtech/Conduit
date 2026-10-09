package openapi

import (
	"fmt"
)

// DocsHTML renders the interactive API documentation page using Scalar.
func DocsHTML(specURL string, title string) []byte {
	if title == "" {
		title = "Conduit API Reference"
	}
	html := fmt.Sprintf(`<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>%s</title>
    <link rel="icon" type="image/svg+xml" href="data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 100 100'><text y='.9em' font-size='90'>⚡</text></svg>">
    <style>
      body {
        margin: 0;
        padding: 0;
        box-sizing: border-box;
      }
    </style>
  </head>
  <body>
    <script
      id="api-reference"
      data-url="%s"
      src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"
    ></script>
  </body>
</html>
`, title, specURL)

	return []byte(html)
}
