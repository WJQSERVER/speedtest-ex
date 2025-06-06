// 你可以把这个放在你的 web 包或者一个工具包里
package web

import (
	"bufio"
	"net"
	"net/http"
	"sync"

	"github.com/infinite-iroha/touka"
)

// safeResponseWriter 是一个线程安全的 touka.ResponseWriter 包装器。
// 它通过互斥锁确保所有对底层 ResponseWriter 的访问都是同步的。
type safeResponseWriter struct {
	touka.ResponseWriter // 内嵌原始的 ResponseWriter
	mu                   sync.Mutex
}

// NewSafeResponseWriter 创建一个新的线程安全包装器。
func NewSafeResponseWriter(w touka.ResponseWriter) *safeResponseWriter {
	return &safeResponseWriter{
		ResponseWriter: w,
		// mu 会被零值初始化
	}
}

// --- 实现所有 touka.ResponseWriter 的方法，并加上锁 ---

func (s *safeResponseWriter) WriteHeader(statusCode int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ResponseWriter.WriteHeader(statusCode)
}

func (s *safeResponseWriter) Write(b []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.ResponseWriter.Write(b)
}

func (s *safeResponseWriter) Flush() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ResponseWriter.Flush()
}

func (s *safeResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.ResponseWriter.Hijack()
}

func (s *safeResponseWriter) Status() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.ResponseWriter.Status()
}

func (s *safeResponseWriter) Size() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.ResponseWriter.Size()
}

func (s *safeResponseWriter) Written() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.ResponseWriter.Written()
}

// Header 方法也需要被包装
func (s *safeResponseWriter) Header() http.Header {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.ResponseWriter.Header()
}
