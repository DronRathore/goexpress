package goexpress

import (
  "io"
  "os"
  "sync"
)

const (
  _mode = 0644
)

type fileHandler struct {
  offset     int64
  bufferSize int64
  f          *os.File
}

// readIntent returns the chunk that has been read
type readIntent struct {
  data []byte
  err  error
}

// monitor monitors the read thread and help sync
// between consumer and producer
type monitor struct {
  sync.Mutex

  stopChannel chan bool
}

func newFile(url string, bufferSize int64) (*fileHandler, error) {
  fh := &fileHandler{}

  file, err := os.Open(url)
  if err != nil {
    return nil, err
  }

  fh.f = file

  if bufferSize == 0 {
    bufferSize = MaxBufferSize
  }
  // assign the buffersize
  fh.bufferSize = bufferSize
  return fh, nil
}

// Stat returns fileinfo of the file
func (fh *fileHandler) Stat() (os.FileInfo, error) {
  return fh.f.Stat()
}

func (fh *fileHandler) Pipe(resp *response) (bool, error) {
  // create a single value buffered channel
  channel := make(chan *readIntent, 1)
  threadMonitor := &monitor{
    // create a single cap channel
    stopChannel: make(chan bool, 1),
  }
  // spin up a read thread
  go fh.readThread(channel, threadMonitor)
  for {
    // read from the channel
    intent := <-channel
    // if there was an error, check the error and return
    if intent.err != nil {
      // notify thread to stop
      threadMonitor.stopChannel <- true
      if intent.err == io.EOF {
        // flush the last bytes and exit
        resp.WriteBytes(intent.data)
        resp.End()
        return true, nil
      }
      return false, intent.err
    }
    // in case of no error, write the bytes
    err := resp.WriteBytes(intent.data)
    if err != nil {
      // stop the file reader go-routine
      threadMonitor.stopChannel <- true
      return false, err
    }
  }
}

func (fh *fileHandler) readThread(channel chan *readIntent, threadMonitor *monitor) {
  // loop over until we exhaust the stream
  for {
    intent := &readIntent{}
    // create a read buffer
    data := make([]byte, fh.bufferSize)
    n, err := fh.f.ReadAt(data, fh.offset)
    // update the offset and return the struct
    fh.offset += int64(n)
    // trim buffer
    intent.data = data[:n]
    intent.err = err
    select {
    case <-threadMonitor.stopChannel:
      return
    case channel <- intent:
    }
  }
}
