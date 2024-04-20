package main

func (app *app) background(fn func()) {

	app.wg.Add(1)
	go func() {
		defer app.wg.Done()
		defer func() {
			if err := recover(); err != nil {
				app.logger.Print(err)
			}
		}()

		fn()
	}()
}
