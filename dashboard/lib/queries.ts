import { gql } from "@apollo/client";

export const GET_TOPOLOGY = gql`
  query GetTopology($filter: TopologyFilter) {
    getTopology(filter: $filter) {
      nodes {
        id
        type
        name
        namespace
        properties
      }
      edges {
        id
        type
        sourceId
        targetId
      }
      statistics {
        totalNodes
        totalEdges
      }
    }
  }
`;

export const GET_DEPENDENCIES = gql`
  query GetDependencies($resourceId: ID!, $filter: DependencyFilter) {
    getDependencies(resourceId: $resourceId, filter: $filter) {
      root {
        id
        name
        type
      }
      dependencies {
        id
        name
        type
      }
      edges {
        id
        sourceId
        targetId
        type
      }
    }
  }
`;
