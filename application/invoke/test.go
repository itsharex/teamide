package invoke

import (
	"fmt"
	"sync"
	"teamide/application/base"
	"teamide/application/common"
	"teamide/application/model"
)

func InvokeTest(app common.IApplication, test *model.TestModel) (res *common.TestResult, err error) {
	res = &common.TestResult{
		Infos: []*common.TestInfo{},
	}
	if len(test.Steps) == 0 {
		return
	}

	if app.GetLogger() != nil && app.GetLogger().OutDebug() {
		app.GetLogger().Debug("test [", test.Name, "] start")
		// app.GetLogger().Debug("test [", test.Name, "] test:", base.ToJSON(testOne))
	}

	startTime := base.GetNowTime()
	defer func() {
		endTime := base.GetNowTime()
		if app.GetLogger() != nil && app.GetLogger().OutDebug() {
			app.GetLogger().Debug("test [", test.Name, "] end, use:", (endTime - startTime), "ms")
		}
	}()
	threadNumber := test.ThreadNumber
	forNumber := test.ForNumber
	if threadNumber <= 0 {
		threadNumber = 1
	}
	if forNumber <= 0 {
		forNumber = 1
	}
	var wg sync.WaitGroup
	wg.Add(threadNumber * forNumber)
	for threadIndex := 0; threadIndex < threadNumber; threadIndex++ {
		go func(tIndex int) {
			for forIndex := 0; forIndex < forNumber; forIndex++ {
				defer wg.Done()

				info := &common.TestInfo{
					ThreadIndex: tIndex,
					ForIndex:    forIndex,
				}
				info.ThreadName = fmt.Sprint("thread-", info.ThreadIndex)
				info.ForName = fmt.Sprint("for-", info.ForIndex)
				res.Infos = append(res.Infos, info)
				e := invokeTest(app, test, info)
				res.Count++
				if e != nil {
					res.ErrorCount++
				} else {
					res.SuccessCount++
				}
			}
		}(threadIndex)
	}

	wg.Wait()
	if err != nil {
		if app.GetLogger() != nil {
			app.GetLogger().Error("test [", test.Name, "] error:", err)
		}
		return
	}
	return
}

func invokeTest(app common.IApplication, test *model.TestModel, info *common.TestInfo) (err error) {

	if app.GetLogger() != nil && app.GetLogger().OutDebug() {
		app.GetLogger().Debug("test [", test.Name, "] [", info.ThreadName, "] [", info.ForName, "] start")
	}

	startTime := base.GetNowTime()
	defer func() {
		if info.Error != nil {

			if app.GetLogger() != nil {
				app.GetLogger().Error("test [", test.Name, "] [", info.ThreadName, "] [", info.ForName, "] error:", info.Error)
			}
		} else {
			endTime := base.GetNowTime()
			if app.GetLogger() != nil && app.GetLogger().OutDebug() {
				app.GetLogger().Debug("test [", test.Name, "] [", info.ThreadName, "] [", info.ForName, "] end, use:", (endTime - startTime), "ms")
			}
		}
	}()

	var javascript string

	javascript, info.Error = common.GetTestJavascriptByTestStep(app, test)

	if info.Error != nil {
		return info.Error
	}

	var invokeNamespace *common.InvokeNamespace
	invokeNamespace, err = common.NewInvokeNamespace(app)
	if err != nil {
		return
	}

	info.Result, info.Error = invokeJavascript(app, invokeNamespace, javascript)
	if info.Error != nil {
		return info.Error
	}
	return nil
}