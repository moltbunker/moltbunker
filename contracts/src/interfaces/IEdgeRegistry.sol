// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/// @title IEdgeRegistry
/// @author Moltbunker
/// @notice Minimal read-only interface for the BunkerEdgeRegistry, consumed by
///         EDGE-02 (the daemon reverse-tunnel stake gate) and by the daemon Go
///         bindings. The daemon's `EdgeRegistryReader` wraps these two functions
///         so it never has to import the full contract ABI.
interface IEdgeRegistry {
    /// @notice Static, on-chain edge-provider metadata.
    /// @dev Mirrored from BunkerEdgeRegistry so consumers can depend on the
    ///      interface alone. The fixed-size leading fields pack into two slots.
    struct EdgeProviderInfo {
        bytes32 nodeId; // SHA256 of the provider's Ed25519 public key
        bytes32 region; // Geographic region identifier
        uint48 registeredAt; // Block timestamp at registration
        bool active; // Whether the provider is currently registered
        bool frozen; // Whether the provider is frozen (emergency response)
        string endpointURL; // Public TLS-terminating edge endpoint
        bytes tlsPubkeyHash; // Hash of the edge node's TLS public key
    }

    /// @notice Returns true if the provider is registered, active, and not frozen.
    /// @param provider The edge-provider address to check.
    /// @return active True when the address can serve edge traffic.
    function isActiveEdgeProvider(address provider) external view returns (bool active);

    /// @notice Returns the full edge-provider metadata struct.
    /// @param provider The edge-provider address to look up.
    /// @return info The provider's registration metadata.
    function getEdgeProviderInfo(address provider)
        external
        view
        returns (EdgeProviderInfo memory info);
}
