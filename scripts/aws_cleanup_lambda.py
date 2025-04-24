#!/usr/bin/env python3
# CloudSnooze AWS Test Resources Cleanup Lambda
# Copyright 2025 Scott Friedman and CloudSnooze Contributors
# SPDX-License-Identifier: Apache-2.0

import boto3
import datetime
import logging
import os

# Configure logging
logger = logging.getLogger()
logger.setLevel(logging.INFO)

# Set up default values
DEFAULT_MAX_AGE_HOURS = 2
DEFAULT_TAG_KEY = "Purpose"
DEFAULT_TAG_VALUE = "Testing"

def get_env_var(name, default):
    """Get an environment variable or return a default value"""
    return os.environ.get(name, default)

def lambda_handler(event, context):
    """
    Lambda function to clean up test resources in AWS
    
    Identifies resources tagged for testing that are older than the specified
    maximum age and terminates/deletes them.
    """
    max_age_hours = int(get_env_var("MAX_AGE_HOURS", DEFAULT_MAX_AGE_HOURS))
    tag_key = get_env_var("TAG_KEY", DEFAULT_TAG_KEY)
    tag_value = get_env_var("TAG_VALUE", DEFAULT_TAG_VALUE)
    
    max_age_seconds = max_age_hours * 3600
    current_time = datetime.datetime.now(datetime.timezone.utc)
    
    logger.info(f"Starting cleanup of resources tagged {tag_key}={tag_value} older than {max_age_hours} hours")
    
    # Initialize results
    results = {
        "terminated_instances": [],
        "errors": []
    }
    
    # Clean up EC2 instances
    try:
        ec2 = boto3.resource('ec2')
        
        # Find all instances with the Testing tag
        instances = ec2.instances.filter(
            Filters=[
                {'Name': f'tag:{tag_key}', 'Values': [tag_value]},
                {'Name': 'instance-state-name', 'Values': ['pending', 'running', 'stopping', 'stopped']}
            ]
        )
        
        # Terminate instances older than the max age
        for instance in instances:
            try:
                # Get launch time
                launch_time = instance.launch_time
                age_seconds = (current_time - launch_time).total_seconds()
                
                if age_seconds > max_age_seconds:
                    instance_id = instance.id
                    logger.info(f"Terminating instance {instance_id} - Age: {age_seconds/3600:.2f} hours")
                    
                    # Get tags for logging
                    tags = {tag['Key']: tag['Value'] for tag in instance.tags or []}
                    logger.info(f"Instance {instance_id} tags: {tags}")
                    
                    # Terminate the instance
                    instance.terminate()
                    results["terminated_instances"].append(instance_id)
                    
                    logger.info(f"Successfully requested termination of instance {instance_id}")
                else:
                    logger.info(f"Skipping instance {instance.id} - Age: {age_seconds/3600:.2f} hours")
            except Exception as e:
                error_msg = f"Error processing instance {instance.id}: {str(e)}"
                logger.error(error_msg)
                results["errors"].append(error_msg)
    
    except Exception as e:
        error_msg = f"Error in EC2 cleanup: {str(e)}"
        logger.error(error_msg)
        results["errors"].append(error_msg)
    
    # Log summary
    logger.info(f"Cleanup complete. Terminated {len(results['terminated_instances'])} instances. Errors: {len(results['errors'])}")
    
    return results

# For local testing
if __name__ == "__main__":
    print("Testing AWS cleanup lambda locally")
    result = lambda_handler({}, None)
    print(f"Result: {result}")