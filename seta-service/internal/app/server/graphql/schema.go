package graphql

var SchemaString = `
	schema {
		query: Query
		mutation: Mutation
	}
	type User {
		userId: ID!
		username: String!
		email: String!
		role: ROLE!
	}
	enum ROLE {
		manager
		member
	}
	type AuthPayload {
		token: String!
		user: User!
	}
	type Query {
		fetchUsers: [User!]!
	}
	type Mutation {
		createUser(username: String!, email: String!, password: String!, role: ROLE!): User!
		login(email: String!, password: String!): AuthPayload!
		logout: Boolean!
	}
`
