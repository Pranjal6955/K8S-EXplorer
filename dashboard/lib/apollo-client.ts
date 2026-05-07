import { ApolloClient, InMemoryCache, HttpLink } from "@apollo/client";

const httpLink = new HttpLink({
    uri: process.env.NEXT_PUBLIC_GRAPHQL_URL || "http://localhost:8080/query",
});

export const client = new ApolloClient({
    link: httpLink,
    cache: new InMemoryCache(),
});
