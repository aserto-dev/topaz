package directory_test

import data.directory as dir

test_user_not_string if {
	res := dir.ds_user(123)
	res == ""
}
