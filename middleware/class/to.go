package class

func ElementStatusFromSql(key string) ClassElementStatus {
	return ClassElementStatus(ClassStatus_value[key])
}

//goland:noinspection GoNameStartsWithPackageName
func ClassStatusFromSql(key string) ClassStatus {
	return ClassStatus(ClassStatus_value[key])
}
