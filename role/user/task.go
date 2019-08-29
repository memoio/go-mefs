package user

import (
	"context"
	"errors"
	"sync"

	pb "github.com/memoio/go-mefs/role/user/pb"
)

var (
	ErrJobIsNil = errors.New("job is nil")
)

// TQuene is a basic priority queue.
type TQuene interface {
	// Push adds the ele
	Push(TaskInfo)
	// Pop returns the highest priority Elem in PQ.
	Pop() TaskInfo
	// Len returns the number of elements in the PQ.
	Len() int
	// Update `fixes` the PQ.
	Update(index int)

	// TODO  explain why this interface should not be extended
	// It does not support Remove. This is because...
}

//TaskQueue ...
//TODO: 待实现
type TaskQueue struct {
	mu        sync.Mutex
	taskQueue TQuene
}

//TaskState 任务状态
type TaskState int32

const (
	//Pending 悬置，待启动
	Pending TaskState = iota
	//Running 运行中
	Running
	//Paused 停顿中
	Paused
	//Completed 已完成
	Completed
	//Error 已出错
	Error
)

//TaskType 任务类型
type TaskType int32

const (
	//UnKnown 未知类型
	UnKnown TaskType = iota
	//Download 下载对象
	DownloadState
	//Upload  上传对象
	UploadState
	//Copy 从一个Bucket复制到另一个Bucket
	Copy
	// Share 从一个Bucket分享对象
	Share
)

//TaskInfo 用于异步执行任务，未来可考虑设计任务队列
// Task两种方案
// A，构造好Job，如Upload，Dowload等，存在Task里，运行就启动，暂停就用某种方式堵塞
// B，Task里指明任务类型，及UserID等所需参数，Start的时候构造出临时Job运行
// 暂停的时候保存好状态（如已上传的index等），将Job释放，重新Start再构建Job运行
type TaskInfo struct {
	LfsService *LfsService
	Object     *pb.ObjectInfo
	BucketID   int32
	Job        Job
	Typ        TaskType  //任务类型
	State      TaskState //任务状态
	Priority   int       //任务的优先级
	Err        error     //如出错，存储错误消息等待获取
}

//Job 具体的工作接口，目前只实现Start，断点续传等后续再做
type Job interface {
	Start(context.Context) error  //启动Job
	Stop(context.Context) error   //停止Job
	Cancel(context.Context) error //取消Job
	Done()                        //回调通知任务完成
	Info() (interface{}, error)   //获取Job信息及状态
}

//NewTask 为Job新建一个任务
func NewTask(typ TaskType, priority int) (*TaskInfo, error) {
	return &TaskInfo{
		Typ:      typ,
		State:    Pending,
		Priority: priority,
	}, nil
}

func (t *TaskInfo) Start(ctx context.Context) error {
	t.State = Completed
	return nil
}

func (t *TaskInfo) Stop(ctx context.Context) error {
	t.State = Paused
	return nil
}

func (t *TaskInfo) Cancel(ctx context.Context) error {
	t.State = Completed
	return nil
}
