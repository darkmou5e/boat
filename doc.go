/*

- All functions that working with an already started transaction panics if
  something goes wrong. Use a wrapper function "DontPanic" (will be implemented soon)
  to catch it. Also use "defer tx.Recover()"" after start a new transaction and
  tx.Commit() imediately after a transaction related code.

  Something like that:

  func myFunc() {
    ...
    tx, _ := db.Begin() // start a transaction
    defer tx.Rollback()

    ... some transaction related code here ...

    tx.Commit()
  }

  DontPanic(myFunc)


*/

package boat
