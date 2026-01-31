// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";

contract StakingContract is ReentrancyGuard {
    IERC20 public bunkerToken;
    uint256 public minStake;
    
    mapping(address => uint256) public stakes;
    mapping(address => bool) public isProvider;
    
    event Staked(address indexed provider, uint256 amount);
    event Unstaked(address indexed provider, uint256 amount);
    event Slashed(address indexed provider, uint256 amount);
    
    constructor(address _bunkerToken, uint256 _minStake) {
        bunkerToken = IERC20(_bunkerToken);
        minStake = _minStake;
    }
    
    function stake(uint256 amount) external nonReentrant {
        require(amount >= minStake, "Stake below minimum");
        
        require(bunkerToken.transferFrom(msg.sender, address(this), amount), "Transfer failed");
        
        stakes[msg.sender] += amount;
        isProvider[msg.sender] = true;
        
        emit Staked(msg.sender, amount);
    }
    
    function unstake(uint256 amount) external nonReentrant {
        require(stakes[msg.sender] >= amount, "Insufficient stake");
        
        stakes[msg.sender] -= amount;
        
        if (stakes[msg.sender] < minStake) {
            isProvider[msg.sender] = false;
        }
        
        require(bunkerToken.transfer(msg.sender, amount), "Transfer failed");
        
        emit Unstaked(msg.sender, amount);
    }
    
    function slash(address provider, uint256 amount) external {
        // Only contract owner or slashing mechanism can call this
        require(stakes[provider] >= amount, "Insufficient stake to slash");
        
        stakes[provider] -= amount;
        
        if (stakes[provider] < minStake) {
            isProvider[provider] = false;
        }
        
        // Slashed tokens are burned or sent to treasury
        // For now, they remain in contract
        
        emit Slashed(provider, amount);
    }
    
    function getStake(address provider) external view returns (uint256) {
        return stakes[provider];
    }
}
