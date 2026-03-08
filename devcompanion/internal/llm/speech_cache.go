package llm

import "sync"

// SpeechCache は LLM レスポンスをメモリ内（FIFO）でキャッシュする。
type SpeechCache struct {
	mu    sync.RWMutex
	cache map[string]string // キー: "{event}:{mood}", 値: キャッシュテキスト
	queue []string          // FIFO 順序管理（古い順）
}

const maxCacheSize = 20

// NewSpeechCache は SpeechCache を初期化する。
func NewSpeechCache() *SpeechCache {
	return &SpeechCache{
		cache: make(map[string]string),
		queue: make([]string, 0, maxCacheSize),
	}
}

// Get はキャッシュから値を取得する。
// 返り値: (値, ヒット判定)
func (sc *SpeechCache) Get(key string) (string, bool) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	val, hit := sc.cache[key]
	return val, hit
}

// Put はキャッシュに値を登録する。
// 容量超過時は FIFO 順序で古い順に削除する。
func (sc *SpeechCache) Put(key string, value string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	// キーがすでに存在する場合、既存の値を更新
	// （FIFO キューからは削除しない）
	if _, exists := sc.cache[key]; exists {
		sc.cache[key] = value
		return
	}

	// 新規キーの場合、キュー追加
	sc.queue = append(sc.queue, key)
	sc.cache[key] = value

	// 容量超過時は最古のキーを削除
	if len(sc.queue) > maxCacheSize {
		oldest := sc.queue[0]
		sc.queue = sc.queue[1:]
		delete(sc.cache, oldest)
	}
}

// Size はキャッシュ内のエントリ数を返す。
func (sc *SpeechCache) Size() int {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	return len(sc.cache)
}
