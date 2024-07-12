package testutil

import "testing/fstest"

type TestFiles map[string]string

func CreateFS(files TestFiles) fstest.MapFS {
	mfs := fstest.MapFS{}

	for path, content := range files {
		mfs[path] = &fstest.MapFile{Data: []byte(content)}
	}

	return mfs
}
