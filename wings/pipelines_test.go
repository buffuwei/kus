package wings

import (
	"buffuwei/kus/tools"
	"testing"
)

func TestAppPipelines(tt *testing.T) {
	asset := tools.GetConfig().GetSelectedAsset()
	wsp := &asset.Wingsplatform
	ps := AppPipelines("content-server", wsp, 10)
	for _, p := range ps {
		tt.Logf("Pipeline: %+v\n", p)
	}
}
