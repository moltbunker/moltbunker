// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";

contract PaymentContract is ReentrancyGuard {
    IERC20 public bunkerToken;
    
    struct Reservation {
        address requester;
        uint256 amount;
        uint256 duration;
        uint256 startTime;
        bool active;
    }
    
    mapping(bytes32 => Reservation) public reservations;
    mapping(bytes32 => address[]) public selectedProviders;
    
    event ReservationCreated(bytes32 indexed reservationId, address indexed requester, uint256 amount, uint256 duration);
    event ProvidersSelected(bytes32 indexed reservationId, address[] providers);
    event PaymentReleased(bytes32 indexed reservationId, address indexed provider, uint256 amount);
    
    constructor(address _bunkerToken) {
        bunkerToken = IERC20(_bunkerToken);
    }
    
    function reserveRuntime(uint256 amount, uint256 duration) external nonReentrant returns (bytes32) {
        require(amount > 0, "Amount must be greater than 0");
        require(duration > 0, "Duration must be greater than 0");
        
        bytes32 reservationId = keccak256(abi.encodePacked(msg.sender, block.timestamp, amount, duration));
        
        reservations[reservationId] = Reservation({
            requester: msg.sender,
            amount: amount,
            duration: duration,
            startTime: 0,
            active: true
        });
        
        // Transfer BUNKER tokens to escrow
        require(bunkerToken.transferFrom(msg.sender, address(this), amount), "Transfer failed");
        
        emit ReservationCreated(reservationId, msg.sender, amount, duration);
        
        return reservationId;
    }
    
    function selectProviders(bytes32 reservationId, address[] calldata providers) external {
        require(reservations[reservationId].active, "Reservation not active");
        require(providers.length == 3, "Must select exactly 3 providers");
        
        reservations[reservationId].startTime = block.timestamp;
        selectedProviders[reservationId] = providers;
        
        emit ProvidersSelected(reservationId, providers);
    }
    
    function releasePayment(bytes32 reservationId, address provider, uint256 amount) external nonReentrant {
        require(reservations[reservationId].active, "Reservation not active");
        require(amount > 0, "Amount must be greater than 0");
        
        // Verify provider is selected
        address[] memory providers = selectedProviders[reservationId];
        bool isSelected = false;
        for (uint i = 0; i < providers.length; i++) {
            if (providers[i] == provider) {
                isSelected = true;
                break;
            }
        }
        require(isSelected, "Provider not selected");
        
        // Release payment
        require(bunkerToken.transfer(provider, amount), "Transfer failed");
        
        emit PaymentReleased(reservationId, provider, amount);
    }
}
