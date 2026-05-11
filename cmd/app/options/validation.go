package options

import "errors"

func (o *Options) Validate() error {
	var err error
	err = errors.Join(err, o.InsecureServing.Validate())
	err = errors.Join(err, o.Database.Validate())
	err = errors.Join(err, o.Cache.Validate())
	err = errors.Join(err, o.JWT.Validate())
	err = errors.Join(err, o.GithubOAuth2.Validate())
	err = errors.Join(err, o.Email.Validate())
	err = errors.Join(err, o.Attachment.Validate())
	return err
}
