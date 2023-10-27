#!/bin/bash
# SPDX-License-Identifier: BUSL-1.1
#
# Copyright (C) 2023, Berachain Foundation. All rights reserved.
# Use of this software is govered by the Business Source License included
# in the LICENSE file of this repository and at www.mariadb.com/bsl11.
#
# ANY USE OF THE LICENSED WORK IN VIOLATION OF THIS LICENSE WILL AUTOMATICALLY
# TERMINATE YOUR RIGHTS UNDER THIS LICENSE FOR THE CURRENT AND ALL OTHER
# VERSIONS OF THE LICENSED WORK.
#
# THIS LICENSE DOES NOT GRANT YOU ANY RIGHT IN ANY TRADEMARK OR LOGO OF
# LICENSOR OR ITS AFFILIATES (PROVIDED THAT YOU MAY USE A TRADEMARK OR LOGO OF
# LICENSOR AS EXPRESSLY REQUIRED BY THIS LICENSE).
#
# TO THE EXTENT PERMITTED BY APPLICABLE LAW, THE LICENSED WORK IS PROVIDED ON
# AN “AS IS” BASIS. LICENSOR HEREBY DISCLAIMS ALL WARRANTIES AND CONDITIONS,
# EXPRESS OR IMPLIED, INCLUDING (WITHOUT LIMITATION) WARRANTIES OF
# MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE, NON-INFRINGEMENT, AND
# TITLE.

sh ./reset-temp.sh

# Stop the old one
docker-compose stop
docker-compose kill

# Run docker-compose in detached mode
docker-compose up -d

# Wait a second
sleep 1

# Initialize the 4 node network
sh ./network-init-4.sh

# Start each node
docker exec -d -it polard-node0 bash -c ./scripts/seed-start.sh 
docker exec -d -it polard-node1 bash -c ./scripts/seed-start.sh 
docker exec -d -it polard-node2 bash -c ./scripts/seed-start.sh 
docker exec -d -it polard-node3 bash -c ./scripts/seed-start.sh 


# Wait for all background processes to finish
wait