//https://studygolang.com/articles/12333
//先抄一遍，模仿写法
package pool

import (
	"errors"
	"io"
	"sync"
	"time"
)

var (
	ErrInvalidConfig = errors.New("invalid pool config")
	ErrPoolClosed    = errors.New("pool closed")
)

type factory func() (io.Closer, error)

type Pool interface {
	Acquire() (io.Closer, error)    // 获取资源
	Release(closer io.Closer) error // 释放资源
	Close(closer io.Closer) error   // 关闭资源
	Shutdown() error                // 关闭池
}

type GenericPool struct {
	sync.Mutex
	pool        chan io.Closer
	maxOpen     int  // 池中最大资源数
	numOpen     int  // 当前池中资源数
	minOpen     int  // 池中最少资源数
	closed      bool // 池是否已关闭
	maxLifetime time.Duration
	factory     factory
}

func NewGenericPool(minOpen, maxOpen int, maxLifetime time.Duration, factory factory) (*GenericPool, error) {
	if maxOpen <= 0 || minOpen > maxOpen {
		return nil, ErrInvalidConfig
	}
	p := &GenericPool{
		maxOpen:     maxOpen,
		minOpen:     minOpen,
		maxLifetime: maxLifetime,
		factory:     factory,
		pool:        make(chan io.Closer, maxOpen),
	}

	for i := 0; i < minOpen; i++ {
		closer, err := factory()
		if err != nil {
			continue
		}
		p.numOpen++
		p.pool <- closer
	}
	return p, nil
}

func (p *GenericPool) Acquire() (io.Closer, error) {
	if p.closed {
		return nil, ErrPoolClosed
	}

	for {
		closer, err := p.getOrCreate()
		if err != nil {
			return nil, err
		}
		// todo maxLifetime处理
		return closer, nil
	}

	return nil, nil
}

func (p *GenericPool) Release(closer io.Closer) error {
	if p.closed {
		return ErrPoolClosed
	}
	p.Lock()
	defer p.Unlock()
	p.pool <- closer
	return nil
}

func (p *GenericPool) getOrCreate() (io.Closer, error) {
	select {
	case closer := <-p.pool:
		return closer, nil
	default:
	}
	p.Lock()
	defer p.Unlock()
	// 应该想表达这个意思 如果开启的数量大于最大值 就不开启了 而是复用之前已释放的
	// 这里是否合理 需要再商榷一下, 取其精华去其糟粕
	if p.numOpen >= p.maxOpen {
		closer := <-p.pool
		return closer, nil
	}

	//新建连接
	closer, err := p.factory()
	if err != nil {
		p.Unlock()
		return nil, err
	}
	p.numOpen++
	return closer, nil
}

// 关闭某个资源
func (p *GenericPool) Close(closer io.Closer) error {
	p.Lock()
	defer p.Unlock()
	err := closer.Close()
	if err != nil {
		p.numOpen--
	}
	return nil
}

// 关闭连接池 释放所有资源
func (p *GenericPool) Shutdown() error {
	if p.closed {
		return ErrPoolClosed
	}
	p.Lock()
	close(p.pool)
	for closer := range p.pool {
		closer.Close()
		p.numOpen--
	}
	p.closed = true
	p.Unlock()
	return nil
}
