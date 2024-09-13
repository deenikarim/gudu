package gudu

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
)

const contentType = "Content-Type"

// Response struct holds the http.ResponseWriter and a map of headers
type Response struct {
	Writer  http.ResponseWriter
	Headers http.Header
}

// NewResponse Initializes a new Response object.
func (g *Gudu) NewResponse() *Response {
	return &Response{
		Headers: make(http.Header),
	}
}

// WriteJSON sets the content type to JSON, marshals the data,
// and sends the response
func (g *Gudu) WriteJSON(w http.ResponseWriter, statusCode int, data interface{}, headers ...http.Header) error {

	// Marshal the data into JSON format
	content, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	if len(headers) > 0 {
		for key, value := range headers[0] {
			w.Header()[key] = value
		}
	}

	// calling the header method here to add a single header
	w.Header().Set(contentType, "application/json")

	// Write the HTTP status code to the response
	w.WriteHeader(statusCode)

	// Write the response content
	_, err = w.Write(content)
	if err != nil {
		return err
	}

	return nil
}

// SetResponseWriter sets the http.ResponseWriter for the Response object
func (r *Response) SetResponseWriter(w http.ResponseWriter) *Response {
	r.Writer = w
	return r
}

// Header Sets a single header.
func (r *Response) Header(key, value string) *Response {
	r.Headers.Set(key, value)
	return r
}

// WithHeaders sets multiple headers at once by iterating over the
// provided map and adding each header to the Headers map.
func (r *Response) WithHeaders(headers http.Header) *Response {
	for key, values := range headers {
		for _, value := range values {
			r.Headers.Add(key, value)
		}
	}
	return r
}

// Send writes all headers and the content to the response.
// It sets the status code and then writes the content.
func (r *Response) Send(content []byte, statusCode int) error {

	for key, values := range r.Headers {
		for _, value := range values {
			r.Writer.Header().Add(key, value)
		}
	}

	// Write the HTTP status code to the response
	r.Writer.WriteHeader(statusCode)
	// Write the response content
	_, err := r.Writer.Write(content)
	if err != nil {
		return err
	}

	return nil
}

// JSON method sets the content type to JSON, marshals the data,
// and sends the response
func (r *Response) JSON(data interface{}, statusCode int) error {
	// Marshal the data into JSON format
	content, err := json.Marshal(data)
	if err != nil {
		http.Error(r.Writer, err.Error(), http.StatusInternalServerError)
		return err
	}

	// calling the header method here to add a single header
	r.Header(contentType, "application/json")

	// Send the JSON content with the given status code
	if err := r.Send(content, statusCode); err != nil {
		http.Error(r.Writer, err.Error(), http.StatusInternalServerError)
		return err
	}
	// otherwise return if everything works very well
	return nil
}

// XML method sets the content type to XML, marshals the data,
// and sends the response
func (r *Response) XML(data interface{}, statusCode int) error {
	// Marshal the data into XML format
	content, err := xml.Marshal(data)
	if err != nil {
		http.Error(r.Writer, err.Error(), http.StatusInternalServerError)
		return err
	}

	// calling the header method here to add a single header
	r.Header(contentType, "application/xml")

	// Send the XML content with the given status code
	if err := r.Send(content, statusCode); err != nil {
		http.Error(r.Writer, err.Error(), http.StatusInternalServerError)
		return err
	}
	// otherwise return if everything works very well
	return nil
}

// HTML method sets the content type to HTML and sends the response
func (r *Response) HTML(content string, status int) error {
	r.Header(contentType, "text/html")
	if err := r.Send([]byte(content), status); err != nil {
		http.Error(r.Writer, err.Error(), http.StatusInternalServerError)
		return err
	}
	return nil
}

// Redirect method sends an HTTP redirect to the client
func (r *Response) Redirect(url string, status int) error {
	r.Header("Location", url)
	if err := r.Send(nil, status); err != nil {
		http.Error(r.Writer, err.Error(), http.StatusInternalServerError)
		return err
	}

	return nil
}

// RedirectPermanent method sends a 301 Moved Permanently redirect
func (r *Response) RedirectPermanent(url string) error {
	err := r.Redirect(url, http.StatusMovedPermanently)
	if err != nil {
		http.Error(r.Writer, err.Error(), http.StatusInternalServerError)
		return err
	}

	return nil
}

// RedirectTemporary method sends a 302 Found redirect
func (r *Response) RedirectTemporary(url string) error {
	err := r.Redirect(url, http.StatusFound)
	if err != nil {
		http.Error(r.Writer, err.Error(), http.StatusInternalServerError)
		return err
	}

	return nil
}

// SetCORS sets CORS(Cross-Origin Resource Sharing)headers to allow all origins
func (r *Response) SetCORS() *Response {
	r.Header("Access-Control-Allow-Origin", "*")
	r.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	r.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
	return r
}

// SetCORSWithOrigin sets CORS(Cross-Origin Resource Sharing)headers to allow a specific origin
func (r *Response) SetCORSWithOrigin(origin string) *Response {
	r.Header("Access-Control-Allow-Origin", origin)
	r.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	r.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
	return r
}

// JSONP Sets the content type to JavaScript, wraps the JSON data in a callback
// function, and sends the response.
func (r *Response) JSONP(data interface{}, callback string, statusCode int) error {
	r.Header(contentType, "application/javascript")
	content, err := json.Marshal(data)
	if err != nil {
		http.Error(r.Writer, err.Error(), http.StatusInternalServerError)
		return err
	}

	// Wrap the JSON content in the callback function
	callBackContent := []byte(callback + "(" + string(content) + ");")

	if err := r.Send(callBackContent, statusCode); err != nil {
		http.Error(r.Writer, err.Error(), http.StatusInternalServerError)
		return err
	}

	return nil
}

// DownloadFile method sets headers for downloading a file and
// streams it to the client
func (r *Response) DownloadFile(pathToFile, fileName string, rr *http.Request) error {
	// Open the file specified by filePath
	filePath := path.Join(pathToFile, fileName)
	fileToServe := filepath.Clean(filePath)

	r.Writer.Header().Set("Content-Disposition", "attachment; filename=\""+fileName+"\"")

	http.ServeFile(r.Writer, rr, fileToServe)

	return nil
}

// StreamDownload method uses a callback function to stream data to the client
// as a download
func (r *Response) StreamDownload(callBack func(writer io.Writer), fileName string, headers map[string]string) error {
	r.Writer.Header().Set("Content-Disposition", "attachment; filename=\""+fileName+"\"")

	for key, value := range headers {
		r.Writer.Header().Set(key, value)
	}

	r.Writer.WriteHeader(http.StatusOK)
	// Execute the callback function, passing the ResponseWriter to stream the data
	callBack(r.Writer)

	return nil
}

// File method sets headers for displaying a file in the browser
// and streams it to the client
func (r *Response) File(fileRoad, fileName string, headers map[string]string) error {
	filePath := path.Join(fileRoad, fileName)
	fileToShow := filepath.Clean(filePath)

	file, err := os.Open(fileToShow)
	if err != nil {
		http.Error(r.Writer, "file not found", http.StatusInternalServerError)
		return err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	for key, value := range headers {
		r.Writer.Header().Set(key, value)
	}

	r.Writer.WriteHeader(http.StatusOK)

	if _, err := io.Copy(r.Writer, file); err != nil {
		http.Error(r.Writer, err.Error(), http.StatusInternalServerError)
		return err
	}
	return nil
}

// HandleFileUpload handles file uploads and saves them to the specified directory
func (r *Response) HandleFileUpload(fieldName, uploadDir string, req *http.Request) (string, error) {
	file, fileHeader, err := req.FormFile(fieldName)
	if err != nil {
		return "", err
	}
	defer func(file multipart.File) {
		_ = file.Close()
	}(file)

	filePath := filepath.Join(uploadDir, fileHeader.Filename)
	outFile, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer func(outFile *os.File) {
		_ = outFile.Close()
	}(outFile)

	if _, err := io.Copy(outFile, file); err != nil {
		http.Error(r.Writer, err.Error(), http.StatusInternalServerError)
		return "", err
	}
	return filePath, nil
}

// =================== error response =================

func (r *Response) Error404() {
	r.errorStatus(http.StatusNotFound)
}

func (r *Response) Error500() {
	r.errorStatus(http.StatusInternalServerError)
}

func (r *Response) ErrorUnauthorized() {
	r.errorStatus(http.StatusUnauthorized)
}

func (r *Response) ErrorForbidden() {
	r.errorStatus(http.StatusForbidden)
}

func (r *Response) errorStatus(status int) {
	http.Error(r.Writer, http.StatusText(status), status)
}
