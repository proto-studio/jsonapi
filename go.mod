module proto.zip/studio/jsonapi

go 1.22.3

require proto.zip/studio/validate v0.4.0

require (
	golang.org/x/net v0.15.0 // indirect
	golang.org/x/text v0.13.0 // indirect
)

retract v0.1.0 // Apply is the 3-arg validate v0.4.0-rc.2 shape and does not compile with validate v0.4.0
