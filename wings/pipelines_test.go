package wings

import (
	"buffuwei/kus/tools"
	"fmt"
	"testing"
)

func TestAppPipelines(tt *testing.T) {
	asset := tools.GetConfig().GetSelectedAsset()
	wsp := asset.Wingsplatform
	ps := AppPipelines("content-server", wsp, 10)
	for _, p := range ps {
		tt.Logf("Pipeline: %+v\n", p)
	}
}

func TestDeploy(tt *testing.T) {
	renewAllCookie()

	wsp := tools.GetConfig().GetSelectedAsset().Wingsplatform
	pipelines := AppPipelines("content-cms-server", wsp, 10)
	p := pipelines[0]
	tag := p.Commits.Branch + "-" + p.Commits.CommitId
	fmt.Printf("Tag: %s\n", tag)
	Deploy("content-cms-server", wsp, tag)
}
