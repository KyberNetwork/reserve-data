#  gin-authenticator

## Example
``` go
func main() {
    r := gin.Default()
    allPermissions = []Permission{ReadOnly, ReadAndWrite}
    readOnlyPermissions = []Permission{ReadOnly}
    secretKey = SecrectKeys{
	    ReadOnly:     "123",
	    ReadAndWrite: "345",
    }
    

	mReadOnly, err := Authenticated(readOnlyPermissions, secrets, NewValidateNonceByTime())
	if err != nil {
		return nil, err
    }
    readOnly := r.Group("/readOnly"){
		readOnly.GET("/a", a)
		readOnly.POST("/b", b)
		readOnly.POST("/c", c)
    }
    readOnly.Use(m)


    mAllPermission, err := Authenticated(allPermissions, secrets, NewValidateNonceByTime())
	if err != nil {
		return nil, err
    }
    allP := r.Group("/allP") {
		allP.GET("/a", a)
		allP.POST("/b", b)
		allP.POST("/c", c)
    }
    allP.Use(mAllPermission)

    r.Run(":8080")
}
```