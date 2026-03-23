package options

func (o *Options) Validate() []error {
	var errs []error

	errs = append(errs, o.Server.Validate()...)

	return errs
}
